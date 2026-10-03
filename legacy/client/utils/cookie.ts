import * as cookie from 'js-cookie'

//存Token
function set(key:string,token:string){

    cookie.set(key,token);
}
//取Token
function get(key){
    
    return cookie.get(key);
}
// 移除Token
function remove(key){
    cookie.remove(key);
}

export const JSCookie = {
    set,
    get,
    remove
}