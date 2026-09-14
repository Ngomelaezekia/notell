INSERT INTO packages (id, name, description, currency, monthly_price, active)
VALUES
 ('pkg_starter','Starter','30 minutes of live streaming per day with metered viewer delivery.','usd',999,TRUE),
 ('pkg_creator','Creator','Higher streaming allowance with recording and increased viewer delivery.','usd',1999,TRUE),
 ('pkg_pro','Pro','High-volume streaming package with expanded viewer delivery and storage.','usd',4999,TRUE),
 ('pkg_247','24/7 Channel','Continuous broadcasting with explicit viewer-delivery and storage allowances; overage is metered.','usd',14999,TRUE)
ON CONFLICT (name) DO UPDATE SET description=EXCLUDED.description,currency=EXCLUDED.currency,monthly_price=EXCLUDED.monthly_price,active=EXCLUDED.active,updated_at=NOW();

INSERT INTO package_allowances (id, package_id, metric, included_quantity, overage_unit_price)
VALUES
 ('allow_starter_stream','pkg_starter','stream_minutes',900,0),
 ('allow_starter_viewers','pkg_starter','viewer_minutes',150000,1000),
 ('allow_starter_storage','pkg_starter','storage_gb_months',10,100000),
 ('allow_creator_stream','pkg_creator','stream_minutes',3600,0),
 ('allow_creator_viewers','pkg_creator','viewer_minutes',500000,900),
 ('allow_creator_storage','pkg_creator','storage_gb_months',50,80000),
 ('allow_pro_stream','pkg_pro','stream_minutes',10800,0),
 ('allow_pro_viewers','pkg_pro','viewer_minutes',2000000,700),
 ('allow_pro_storage','pkg_pro','storage_gb_months',200,60000),
 ('allow_247_stream','pkg_247','stream_minutes',44640,0),
 ('allow_247_viewers','pkg_247','viewer_minutes',5000000,600),
 ('allow_247_storage','pkg_247','storage_gb_months',1000,50000)
ON CONFLICT (package_id, metric) DO UPDATE SET included_quantity=EXCLUDED.included_quantity,overage_unit_price=EXCLUDED.overage_unit_price;
