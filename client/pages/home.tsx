import * as React from 'react'
import GAProvider from './../ga-provider'
import {Layout} from './../layout'
import {ToolItem} from './../component'

export default class Home extends React.Component{

    static contextType = GAProvider.context;

    static title = "我是首页啊"
    // 数据获取
    static async initData(params:any) {

        let list = [
            {title:"计算器",href:"/calculator",color:"#c97b4a"},
            {title:"加密解密",href:"/crypto",color:"#c97b4a"},
            {title:"格式化",href:"/prettier",color:"#c97b4a"},
            {title:"进制转换",href:"/hexconvert",color:"#c97b4a"},
            {title:"时间戳转换",href:"/moment",color:"#c97b4a"},
            {title:"终端",href:"/xterm",color:"#c97b4a"},
            {title:"谷歌翻译",href:"/translate",color:"#c97b4a"},
            {title:"Protobuf",href:"/protobuf",color:"#c97b4a"},
            {title:"RGB",href:"/rgb",color:"#c97b4a"},
            {title:"PDF",href:"/pdf",color:"#c97b4a"},
            {title:"Canvas",href:"/canvas",color:"#c97b4a"},
        ]
        return {content:"首页",list:list}
    }
    constructor(props:any){
        super(props)
    }

    render(){
        const { content,list } = this.context

        return (
            <Layout>
                <div className="container home">
                    <div className="tools">
                        {list?list.map((item,index)=>{
                            return <ToolItem
                                {...item}
                                key={"toolitem"+index}
                                ></ToolItem>
                        }):null}
                    </div>
                </div>
            </Layout>
        )
    }
}