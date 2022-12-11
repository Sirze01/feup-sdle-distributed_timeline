import { useState } from "react";

function useToken() {
    const getToken = () => {
        const tokenString = sessionStorage.getItem('token');
        if(tokenString){
            return tokenString;
        }
        else{
            return undefined;
        }
    
    };
    
    const [token, setToken] = useState(getToken());
    
    const saveToken = userToken => {
        if(userToken === undefined){
            sessionStorage.removeItem('token');
        } else {
            sessionStorage.setItem('token', userToken);
            setToken(userToken.token);
        }
    };
    
    const clearToken = ()=>{
        sessionStorage.removeItem('token');
        setToken();
    }
    return {
        token,
        setToken: saveToken,
        clearToken:clearToken
    }
}

export default useToken;