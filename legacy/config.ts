
// 特殊页面 独立打包
import Prettier from './client/pages/prettier/prettier'
// 特殊页面 独立打包
import Home from './client/pages/home/home'
import {RouteItem} from './interface'

// 客户端与服务端都需要的路由控制 url 路由 key 对应ejs页面 special 是否需要单独组件渲染隔离 component 组件
export const routes:RouteItem[] = [    
    {
        url:"/",
        key:"home",
        special:true,
        component:Home
    },{
        url:"/calculator",
        key:"calculator",
        special:false
    },{
        url:"/crypto",
        key:"crypto",
        special:false
    },{
        url:"/prettier",
        key:"prettier",
        special:true,
        component:Prettier
    },{
        url:"/hexconvert",
        key:"hexconvert",
        special:false
    },{
        url:"/rgb",
        key:"rgb",
        special:false
    },{
        url:"/moment",
        key:"moment",
        special:false
    },{
        url:"/protobuf",
        key:"protobuf",
        special:false
    },{
        url:"/rmbconvert",
        key:"rmbconvert",
        special:false
    }
]

export default {
    staticHost:'',
}