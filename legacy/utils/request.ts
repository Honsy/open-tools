var axios = require('axios')
var config = require('../config')
var KoaCache = require('./cache')
var cookie = require('./../client/utils/cookie')


// 是否是客户端使用
var isClient = true
if (typeof document != "undefined") {
  isClient = true
}else{
  isClient = false
}

// 创建实例时设置配置的默认值
var instance = axios.create({
    baseURL: config.BASE_URL,
});

// 修改默认值 如果是客户端从js-cookie中获取 如果是服务端KoaCache中获取
// 添加请求拦截器
instance.interceptors.request.use(function (config) {
  config.headers = {
    "AUTH-TOKEN":isClient?cookie.UCookie.getToken()?cookie.UCookie.getToken():null:KoaCache.get("_t")?KoaCache.get("_t"):null
  }
  // 在发送请求之前做些什么
    return config;
  }, function (error) {
    // 对请求错误做些什么
    return Promise.reject(error);
});

// 添加响应拦截器
instance.interceptors.response.use(function (response) {
    var data = response.data
    // 1001 长时间显示错误 目前只针对邮件的excel上传
    if (data.code === 1001) {
      // message.error(data.message,10)
      return Promise.reject(data.message);
    }

    if (data.code !== 200) {
        // message.error(data.message,2.5)
        return Promise.reject(data.message);
    }

    // 清楚本地登录状态
    if(data.code === 40010||data.code===40011||data.code===40012||data.code===40013 ){

    }

    // 对响应数据做点什么
    return response;
  }, function (error) {
    // 对响应错误做点什么
    // message.error('网络错误，请稍后再试！',2.5)
    return Promise.reject(error);
});

module.exports = instance