import {
 createContext,
 useContext
} from "react";

import {
 useUser
} from "../hooks/useUser";


const UserContext =
createContext(null);



export function UserProvider({
 children
}){


 const userState =
 useUser();



 return (

  <UserContext.Provider
    value={userState}
  >

    {children}

  </UserContext.Provider>

 );

}



export function useCurrentUser(){

 const context =
 useContext(UserContext);


 if(!context){

  throw new Error(
   "useCurrentUser must be inside UserProvider"
  );

 }


 return context;

}