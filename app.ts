var Koa = require('koa');
var app = new Koa();
var views = require('koa-views')
var path = require('path')
var ssr = require('./utils/ssr')
var utils = require('./utils')
var koaStatic = require('koa-static');
var client = require('./utils/client')
var logger = require('./utils/logger')
var Router = require('koa-router');
var router = new Router();

// 加载模板引擎
app.use(views(path.join(__dirname, './templates'), {extension: 'ejs'}))
// 静态文件访问
app.use(koaStatic(path.join(__dirname,'./static')));
// 错误处理
app.on('error', (err:any) =>  {logger.errlogger.error('ERR',err)});
// Page Router
router.get('/',async (ctx) =>{
    let data = await client.reactComponentHandle(ctx)
    await ctx.render('home', render(data.title,data.data,data.ssrData.html,data.state))
})

// Page Router
router.get('/calculator',async (ctx) =>{
    let data = await client.reactComponentHandle(ctx)
    await ctx.render('calculator', render(data.title,data.data,data.ssrData.html,data.state))
})

// Page Router
router.get('/crypto',async (ctx) =>{
    let data = await client.reactComponentHandle(ctx)
    await ctx.render('crypto', render(data.title,data.data,data.ssrData.html,data.state))
})

// Spec Page Router
router.get('/prettier',async (ctx) =>{
    let data = await client.specReactComponentHandle(ctx)
    await ctx.render('prettier', render(data.title,data.data,data.ssrData.html,data.state))
})

// Page Router
router.get('/hexconvert',async (ctx) =>{
    let data = await client.reactComponentHandle(ctx)
    await ctx.render('hexconvert', render(data.title,data.data,data.ssrData.html,data.state))
})

// Page Router
router.get('/rgb',async (ctx) =>{
    let data = await client.reactComponentHandle(ctx)
    await ctx.render('rgb', render(data.title,data.data,data.ssrData.html,data.state))
})

app.use(router.routes());


// 渲染配置
function render(title:string,initData:any,body:any,state:any){
    let head = utils.templateHead
    let foot = utils.templateFoot
    return {
        title,
        head:head(),
        foot:foot(),
        initData:'<script id="js-data" type="text/json">'+ JSON.stringify(initData) +'</script>',
        initState:'<script>window.STATE_FROM_SERVER = '+JSON.stringify(state)+'</script>',
        body
    }
}


app.listen(3000);