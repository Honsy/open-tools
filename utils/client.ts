var logger = require('./logger')
var ssr = require('./ssr')

// 如果需要组件匹配 使用此匹配
import { matchRoutes } from "react-router-config";
import routes from './../client/router.config'
// redux同构
import { createServerStore } from './../client/redux/store'

// 常规RouterConfig获取数据模板
async function reactComponentHandle(ctx:any){
    if(ctx.URL.pathname.indexOf('.')>-1) { 
        return false;
    }
    // 组件
    let branch =  matchRoutes(routes,ctx.URL.pathname);
    let component:any
    // 路由参数
    let params:any
    if (branch.length != 0) {
        component = branch[0].route.component;
        params = branch[0].match.params
    }else{
        branch = matchRoutes(routes,'/404');
        component = branch[0].route.component;
    }
    // redux state
    let state:any = {}
    // 数据预取data
    let data:any = {};
    try {
        data = await component.initData(params)

    } catch (error) {
        logger.errlogger.error('获取数据错误 ERR',error)
        await ctx.render('error',{message:error})
    }
    // 状态
    let store = createServerStore({})
    // 标题
    let normalTitle = component.title?component.title:"通用标题"
    let dynamicTitle = data?data.hasOwnProperty('dynamicTitle')?data.dynamicTitle:null:null
    let title = dynamicTitle?dynamicTitle:normalTitle
    let ssrData = ssr({componentName:"SPA",data:{url:ctx.URL.pathname,data:data}},store)

    return {
        title,
        data,
        ssrData,
        state
    }
}

// 特殊非单页独立页面获取组件数据
async function specReactComponentHandle(ctx:any,ownComponent:any){
    if(ctx.URL.pathname.indexOf('.')>-1) { 
        return false;
    }
    // redux state
    let state:any = {}
    // 数据预取data
    let data:any = {};
    let params:any = {};
    let component:any
    component = ownComponent
    try {
        data = await component.initData()

    } catch (error) {
        logger.errlogger.error('获取数据错误 ERR',error)
        await ctx.render('error',{message:error})
    }
    // 状态
    let store = createServerStore({})
    // 标题
    let normalTitle = component.title?component.title:"通用标题"
    let dynamicTitle = data?data.hasOwnProperty('dynamicTitle')?data.dynamicTitle:null:null
    let title = dynamicTitle?dynamicTitle:normalTitle
    let ssrData = ssr({componentName:"SPA",data:{url:ctx.URL.pathname,data:data}},store)
    return {
        title,
        data,
        ssrData,
        state
    }
}


module.exports = {
    reactComponentHandle,
    specReactComponentHandle
}