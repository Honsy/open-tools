import * as React from 'react'
import GAProvider from '../../ga-provider'
import {Layout} from '../../layout'
import {ToolItem} from '../../component'
import BGCanvas2 from '../../component/canvas/bgCanvas2'


export default class Home extends React.Component{

    static contextType = GAProvider.context;

    static title = "我是首页啊"
    // 数据获取
    static async initData(params:any) {

        let list = [
            {title:"计算器",href:"/calculator",color:"#FFB6C1"},
            {title:"加密解密",href:"/crypto",color:"#E6E6FA"},
            {title:"格式化",href:"/prettier",color:"#4169E1"},
            {title:"进制转换",href:"/hexconvert",color:"#5F9EA0"},
            {title:"时间戳转换",href:"/moment",color:"#8FBC8F"},
            {title:"终端",href:"/xterm",color:"#BDB76B"},
            {title:"谷歌翻译",href:"/translate",color:"#FAEBD7"},
            {title:"Protobuf",href:"/protobuf",color:"#E9967A"},
            {title:"RGB",href:"/rgb",color:"#C0C0C0"},
            {title:"PDF",href:"/pdf",color:"#008080"},
            {title:"Canvas",href:"/canvas",color:"#1E90FF"},
        ]
        return {content:"首页",list:list}
    }

    constructor(props:any){
        super(props)
    }

    componentDidMount(){

    }

    render(){
        const { content,list } = this.context

        return (
            <Layout>
                <div className="container home">
                    <BGCanvas2></BGCanvas2>
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