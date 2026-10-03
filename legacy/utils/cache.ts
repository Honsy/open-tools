

let data = {
    
}
/**
 * 缓存
 * @param name 键
 * @param value 值
 */
function set(name,value){
    data[name] = value
}

/**
 * 返回值
 * @param name 键
 */
function get(name){
    return data[name]
}

module.exports = {
    set,
    get
}