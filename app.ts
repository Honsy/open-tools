import Koa = require('koa');
import views = require('koa-views')
import path = require('path')
import koaStatic = require('koa-static');
import { specReactComponentHandle,reactComponentHandle } from './utils/client'
import {errlogger}  from './utils/logger'
import Router = require('koa-router');
import {routes} from './config'
import { templateHead, templateFoot } from './utils';

var app = new Koa();
var router = new Router();

// 加载模板引擎
app.use(views(path.join(__dirname, './templates'), {extension: 'ejs'}))
// 静态文件访问
app.use(koaStatic(path.join(__dirname,'./static')));
// 错误处理
app.on('error', (err:any) =>  {errlogger.error('ERR',err)});
// Page Router
routes.map((item,index)=>{
    router.get(item.url,async (ctx) =>{
        let data = item.special?await specReactComponentHandle(ctx,item.component,item):await reactComponentHandle(ctx)
        await ctx.render(item.key, render(data.title,data.data,data.ssrData.html,data.state))
    })
})

app.use(router.routes());


// 渲染配置
function render(title:string,initData:any,body:any,state:any){
    return {
        title,
        head:templateHead(),
        foot:templateFoot(),
        initData:'<script id="js-data" type="text/json">'+ JSON.stringify(initData) +'</script>',
        initState:'<script>window.STATE_FROM_SERVER = '+JSON.stringify(state)+'</script>',
        body
    }
}


app.listen(3000);