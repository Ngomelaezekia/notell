package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"notell/config"
	"notell/models"
	"notell/observability"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/gorm"
)

type MediaStorage interface { Put(context.Context,string,string,string) error; Delete(context.Context,string) error; Exists(context.Context,string)(bool,error); Open(context.Context,string,string)(io.ReadCloser,string,int64,string,error) }
type mediaPlaybackSigner interface { GeneratePlaybackURL(context.Context,string,time.Duration)(string,error) }
type localMediaStorage struct{}
type s3MediaStorage struct { client *s3.Client; bucket string; signer *B2MediaSigner }
var mediaStorageRegistry struct { sync.RWMutex; storage MediaStorage }
const orphanMediaGracePeriod = time.Hour

func NewMediaStorage(cfg *config.Config) (MediaStorage,error) {
	if strings.EqualFold(cfg.StorageDriver,"local") { return &localMediaStorage{},nil }
	if !strings.EqualFold(cfg.StorageDriver,"b2") { return nil,fmt.Errorf("unsupported STORAGE_DRIVER %q",cfg.StorageDriver) }
	if cfg.B2Endpoint=="" || cfg.B2Bucket=="" || cfg.B2KeyID=="" || cfg.B2ApplicationKey=="" || cfg.B2Region=="" { return nil,errors.New("Backblaze B2 storage configuration is incomplete") }
	awsCfg,err:=awsconfig.LoadDefaultConfig(context.Background(),awsconfig.WithRegion(cfg.B2Region),awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.B2KeyID,cfg.B2ApplicationKey,"")))
	if err!=nil{return nil,fmt.Errorf("failed to initialize B2 credentials: %w",err)}
	client:=s3.NewFromConfig(awsCfg,func(o *s3.Options){o.BaseEndpoint=aws.String(strings.TrimRight(cfg.B2Endpoint,"/"));o.UsePathStyle=true})
	return &s3MediaStorage{client:client,bucket:cfg.B2Bucket,signer:NewB2MediaSigner(client,cfg.B2Bucket)},nil
}
func cleanLocalMediaPath(key string)(string,error){ c:=filepath.Clean(filepath.FromSlash(key)); if c=="."||filepath.IsAbs(c)||c==".."||strings.HasPrefix(c,".."+string(filepath.Separator)){return "",errors.New("invalid local media key")};return filepath.Join(".",c),nil }
func (s *localMediaStorage) Put(_ context.Context,key,src,_ string) error { dst,err:=cleanLocalMediaPath(key);if err!=nil{return err};if err=os.MkdirAll(filepath.Dir(dst),0755);err!=nil{return fmt.Errorf("create local media directory: %w",err)};in,err:=os.Open(src);if err!=nil{return fmt.Errorf("open local media source: %w",err)};defer in.Close();tmp,err:=os.CreateTemp(filepath.Dir(dst),".media-*");if err!=nil{return err};tmpPath:=tmp.Name();defer os.Remove(tmpPath);if _,err=io.Copy(tmp,in);err!=nil{tmp.Close();return fmt.Errorf("copy local media: %w",err)};if err=tmp.Close();err!=nil{return err};if err=os.Chmod(tmpPath,0644);err!=nil{return err};if err=os.Rename(tmpPath,dst);err!=nil{return fmt.Errorf("commit local media: %w",err)};return nil }
func (s *localMediaStorage) Delete(_ context.Context,key string)error{p,e:=cleanLocalMediaPath(key);if e!=nil{return e};e=os.Remove(p);if e!=nil&&!errors.Is(e,os.ErrNotExist){return fmt.Errorf("delete local media: %w",e)};return nil}
func (s *localMediaStorage) Exists(_ context.Context,key string)(bool,error){p,e:=cleanLocalMediaPath(key);if e!=nil{return false,e};_,e=os.Stat(p);if e==nil{return true,nil};if errors.Is(e,os.ErrNotExist){return false,nil};return false,e}
func (s *localMediaStorage) Open(_ context.Context,key,rng string)(io.ReadCloser,string,int64,string,error){p,e:=cleanLocalMediaPath(key);if e!=nil{return nil,"",0,"",e};f,e:=os.Open(p);if e!=nil{return nil,"",0,"",e};info,e:=f.Stat();if e!=nil{f.Close();return nil,"",0,"",e};ct:=mimeTypeForExtension(strings.ToLower(filepath.Ext(p)));start,end,ok:=parseByteRange(rng,info.Size());if rng!=""&&!ok{f.Close();return nil,"",0,"",errors.New("invalid byte range")};if ok{if _,e=f.Seek(start,io.SeekStart);e!=nil{f.Close();return nil,"",0,"",e};return &limitedReadCloser{Reader:io.LimitReader(f,end-start+1),Closer:f},ct,end-start+1,fmt.Sprintf("bytes %d-%d/%d",start,end,info.Size()),nil};return f,ct,info.Size(),"",nil}
func mimeTypeForExtension(e string)string{switch e{case ".jpg",".jpeg":return "image/jpeg";case ".png":return "image/png";case ".webp":return "image/webp";case ".mp3":return "audio/mpeg";case ".mp4":return "video/mp4";case ".mov":return "video/quicktime";default:return "application/octet-stream"}}
func (s *s3MediaStorage) Put(ctx context.Context,key,path,ct string)error{f,e:=os.Open(path);if e!=nil{observability.MediaStorageFailed("put_open",e);return fmt.Errorf("open media for B2 upload: %w",e)};defer f.Close();info,e:=f.Stat();if e!=nil{return e};_,e=s.client.PutObject(ctx,&s3.PutObjectInput{Bucket:aws.String(s.bucket),Key:aws.String(key),Body:f,ContentType:aws.String(ct),ContentLength:aws.Int64(info.Size()),CacheControl:aws.String("private, max-age=31536000, immutable")});if e!=nil{observability.MediaStorageFailed("put",e);return fmt.Errorf("upload media to Backblaze B2: %w",e)};return nil}
func (s *s3MediaStorage) Delete(ctx context.Context,key string)error{_,e:=s.client.DeleteObject(ctx,&s3.DeleteObjectInput{Bucket:aws.String(s.bucket),Key:aws.String(key)});if e!=nil{observability.MediaStorageFailed("delete",e);return fmt.Errorf("delete media from Backblaze B2: %w",e)};return nil}
func (s *s3MediaStorage) Exists(ctx context.Context,key string)(bool,error){_,e:=s.client.HeadObject(ctx,&s3.HeadObjectInput{Bucket:aws.String(s.bucket),Key:aws.String(key)});if e==nil{return true,nil};if s3StatusCode(e)==404{return false,nil};observability.MediaStorageFailed("exists",e);return false,fmt.Errorf("check media in Backblaze B2: %w",e)}
func s3StatusCode(e error)int{type statusCoder interface{HTTPStatusCode()int};var sc statusCoder;if errors.As(e,&sc){return sc.HTTPStatusCode()};return 0}
func (s *s3MediaStorage) Open(ctx context.Context,key,rng string)(io.ReadCloser,string,int64,string,error){in:=&s3.GetObjectInput{Bucket:aws.String(s.bucket),Key:aws.String(key)};if rng!=""{if !validOpenEndedByteRange(rng){return nil,"",0,"",errors.New("invalid byte range")};in.Range=aws.String(rng)};out,e:=s.client.GetObject(ctx,in);if e!=nil{observability.MediaStorageFailed("open",e);return nil,"",0,"",fmt.Errorf("read media from Backblaze B2: %w",e)};ct:="application/octet-stream";if out.ContentType!=nil&&strings.TrimSpace(*out.ContentType)!=""{ct=*out.ContentType};var n int64;if out.ContentLength!=nil{n=*out.ContentLength};cr:="";if out.ContentRange!=nil{cr=*out.ContentRange};return out.Body,ct,n,cr,nil}
func validOpenEndedByteRange(v string)bool{v=strings.TrimSpace(v);if v==""||!strings.HasPrefix(v,"bytes="){return false};p:=strings.Split(strings.TrimPrefix(v,"bytes="),"-");if len(p)!=2{return false};if p[0]==""{if p[1]==""{return false};n,e:=strconv.ParseInt(p[1],10,64);return e==nil&&n>0};a,e:=strconv.ParseInt(p[0],10,64);if e!=nil||a<0{return false};if p[1]==""{return true};b,e:=strconv.ParseInt(p[1],10,64);return e==nil&&b>=a}
func (s *s3MediaStorage) GeneratePlaybackURL(ctx context.Context,path string,d time.Duration)(string,error){if s==nil||s.signer==nil{return "",errors.New("B2 playback signer is not initialized")};return s.signer.GeneratePlaybackURL(ctx,path,d)}
func SetMediaStorage(s MediaStorage){mediaStorageRegistry.Lock();mediaStorageRegistry.storage=s;mediaStorageRegistry.Unlock()}
func DeleteMediaObject(ctx context.Context,key string)error{mediaStorageRegistry.RLock();s:=mediaStorageRegistry.storage;mediaStorageRegistry.RUnlock();if s==nil{return errors.New("media storage is not registered")};return s.Delete(ctx,key)}
type limitedReadCloser struct{io.Reader;Closer io.Closer};func(r *limitedReadCloser)Close()error{return r.Closer.Close()}
func parseByteRange(v string,size int64)(int64,int64,bool){v=strings.TrimSpace(v);if v==""||!strings.HasPrefix(v,"bytes="){return 0,0,false};p:=strings.Split(strings.TrimPrefix(v,"bytes="),"-");if len(p)!=2{return 0,0,false};if p[0]==""{if p[1]==""||size<=0{return 0,0,false};n,e:=strconv.ParseInt(p[1],10,64);if e!=nil||n<=0{return 0,0,false};if n>size{n=size};return size-n,size-1,true};a,e:=strconv.ParseInt(p[0],10,64);if e!=nil||a<0||a>=size{return 0,0,false};b:=size-1;if p[1]!=""{b,e=strconv.ParseInt(p[1],10,64);if e!=nil||b<a{return 0,0,false};if b>=size{b=size-1}};return a,b,true}
func StartMediaReconciler(ctx context.Context,storage MediaStorage,db *gorm.DB){SetMediaStorage(storage);go func(){reconcile:=func(){ReconcileMediaState(db);if s,ok:=storage.(*s3MediaStorage);ok{if e:=s.reconcile(ctx,db);e!=nil{observability.MediaStorageFailed("reconcile",e);log.Printf("media reconciliation failed: %v",e)};if e:=ReconcileDatabaseMedia(ctx,storage,db);e!=nil{observability.MediaStorageFailed("reconcile_db",e);log.Printf("media database availability reconciliation failed: %v",e)}}};reconcile();t:=time.NewTicker(30*time.Minute);defer t.Stop();for{select{case<-ctx.Done():return;case<-t.C:reconcile()}}}()}
func ReconcileDatabaseMedia(ctx context.Context,storage MediaStorage,db *gorm.DB)error{var us []models.Upload;if e:=db.Select("id, filename, post_id").Where("filename <> ''").Find(&us).Error;e!=nil{return fmt.Errorf("list database media: %w",e)};for _,u:=range us{ok,e:=storage.Exists(ctx,MediaObjectKey(u.Filename));if e!=nil{return e};if ok{continue};msg:="physical media object is missing from durable storage";if u.PostID==nil{msg+="; upload must be replaced"};if e=db.Model(&models.MediaMetadata{}).Where("upload_id = ?",u.ID).Updates(map[string]any{"status":"failed","processing_error":msg}).Error;e!=nil{return fmt.Errorf("mark missing media upload=%d: %w",u.ID,e)}};return nil}
func(s *s3MediaStorage)reconcile(ctx context.Context,db *gorm.DB)error{p:=s3.NewListObjectsV2Paginator(s.client,&s3.ListObjectsV2Input{Bucket:aws.String(s.bucket),Prefix:aws.String("uploads/")});cut:=time.Now().Add(-orphanMediaGracePeriod);for p.HasMorePages(){page,e:=p.NextPage(ctx);if e!=nil{return fmt.Errorf("list Backblaze B2 media: %w",e)};for _,o:=range page.Contents{if o.Key==nil||!strings.HasPrefix(*o.Key,"uploads/"){continue};key:=*o.Key;fn:=strings.TrimPrefix(key,"uploads/");var u models.Upload;e=db.Where("filename = ?",fn).First(&u).Error;if e==nil{continue};if !errors.Is(e,gorm.ErrRecordNotFound){return fmt.Errorf("find upload for %q: %w",fn,e)};created:=time.Now();if o.LastModified!=nil{created=*o.LastModified};if created.After(cut){continue};if e=s.Delete(ctx,key);e!=nil{return fmt.Errorf("delete orphan media %q: %w",key,e)}}};return nil}
