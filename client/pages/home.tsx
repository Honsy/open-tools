import * as React from 'react'
import GAProvider from './../ga-provider'
import {Layout} from './../layout'
import {ToolItem} from './../component'


const width = 1344
const height = 600
const rectX = width/2
const rectY = height/2

// 粒子类实现
class Particle{
    x:number
    y:number
    vx:number
    vy:number
    constructor(x:number,y:number){
        this.x = x
        this.y = y
        this.vx = Math.random() * 3;
        this.vy = Math.random() * 2;
    }

    render(ctx:CanvasRenderingContext2D){
        this.x >= rectX? this.x += this.vx:this.x -= this.vx;
        this.y >= rectY? this.y += this.vy:this.y -= this.vy;
        ctx.beginPath();
        ctx.fillStyle = "#FFA500";
        ctx.arc(this.x, this.y, 2, 0, 2 * Math.PI);
        ctx.fill();
    }
}

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

    bgCanvas:React.RefObject<HTMLCanvasElement>
    bgCtx:CanvasRenderingContext2D
    dymRadius:number = 200
    particles:Particle[]
    
    constructor(props:any){
        super(props)
        this.bgCanvas = React.createRef()
        this.bgCtx = null
        this.particles = []
    }

    componentDidMount(){
        // 初始化WebGL上下文
        this.bgCtx = this.bgCanvas.current.getContext("2d")
        window.requestAnimationFrame(this.draw)
        // this.draw()
    }

    draw = () =>{
        // 清空canvase
        let radius = this.dymRadius > 100 ? this.dymRadius-=0.3:this.dymRadius=200
        
        console.log("drw",radius,this.dymRadius)
        this.bgCtx.clearRect(0,0,width,height)
        //  1000 点生成
        // this.bgCtx.clearRect(0,0,width,height)
        // 每帧生成 1000个点
        // 已经生成的点保存 后动态运动 明天写一下
        for (let index = 0; index < 10; index++) {
            // 生成60 - 120之间的随机数字
            let x = -Math.floor(Math.random() * radius)
            // 求对应点的y值
            let y = Math.sqrt(Math.pow(radius,2) - Math.pow(x,2))
            let newX = Math.random() < 0.5 ? x : -x
            let newY = Math.random() < 0.5 ? y : -y
            let particle = new Particle(newX+rectX,newY+rectY)
            this.particles.push(particle)
        }

        this.particles.forEach((particle,index) => {
            let isRemove = Math.pow(Math.abs(particle.x - rectX),2) + Math.pow(Math.abs(particle.y- rectY),2) - Math.pow(300,2)
            if (isRemove>0) {
                this.particles.splice(index,1)
            }
            particle.render(this.bgCtx)
        });

        this.bgCtx.beginPath()
        this.bgCtx.fillStyle = "#363636";
        this.bgCtx.arc(rectX,rectY,radius,0,Math.PI*2,true)
        this.bgCtx.fill()

        window.requestAnimationFrame(this.draw)
    }


    animate = () =>{
        
    }


    render(){
        const { content,list } = this.context

        return (
            <Layout>
                <div className="container home has-background-dark">
                    <canvas width={width} height={height} className="bgCanvas" ref={this.bgCanvas}>

                    </canvas>
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