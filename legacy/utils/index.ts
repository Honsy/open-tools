import config from './../config'

// 模板头部
export function templateHead(){
    var head = 
    '<meta charset="utf-8">'+
    '<meta http-equiv="X-UA-Compatible" content="ie=edge">' +
    '<link rel="shortcut icon" href="/img/favicon.ico">' +
    '<meta name="viewport" content="width=device-width,initial-scale=1,user-scalable=no,viewport-fit=cover">'+
    // '<script type="text/javascript" src="//connect.qq.com/qc_jssdk.js" data-appid="101748432" charset="utf-8"></script>'+
    '<link  rel="stylesheet" type="text/css" href="'+config.staticHost+'/css/index.css"/>' +
    '<link  rel="stylesheet" type="text/css" href="'+config.staticHost+'/css/bulma.css"/>'
    return head
}

// 模板尾部
export function templateFoot(){
    var foot = 
    '<script src="'+config.staticHost+'/js/react-vendors.js"></script>' +
    '<script src="'+config.staticHost+'/js/react.js"></script>'
    return foot
}