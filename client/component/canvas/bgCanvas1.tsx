import React from 'react'

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
    alpha:number = 1
    angle:number
    constructor(x:number,y:number){
        this.x = x
        this.y = y
        this.vx = Math.random() * 6;
        this.vy = Math.random() * 7;
        this.angle = Math.random() * 360
    }

    render(ctx:CanvasRenderingContext2D,alpha:boolean){
        this.x >= rectX? this.x += this.vx * Math.abs(Math.sin(this.angle)):this.x -= this.vx * Math.abs(Math.sin(this.angle));
        this.y >= rectY? this.y += this.vy * Math.abs(Math.cos(this.angle)):this.y -= this.vy * Math.abs(Math.cos(this.angle));


        ctx.beginPath();
        ctx.fillStyle = "rgba(177,17,22,"+(alpha?1:this.alpha-=0.02)+")";
        ctx.arc(this.x, this.y, 1, 0, 2 * Math.PI);
        ctx.fill();
    }
}

export default class bgCanvas1 extends React.Component{

    bgCanvas:React.RefObject<HTMLCanvasElement>
    bgCtx:CanvasRenderingContext2D
    dymRadius:number = 200
    particles:Particle[]
    shapePath:Path2D

    constructor(props:any){
        super(props)
        this.bgCanvas = React.createRef()
        this.bgCtx = null
        this.particles = []
        this.shapePath = new Path2D()
    }

    componentDidMount(){
        // 初始化WebGL上下文
        this.bgCtx = this.bgCanvas.current.getContext("2d")
        // 透明开始Path
        let alphaPath = new Path2D()
        alphaPath.moveTo(rectX+250,rectY)
        // 第一个控制点 第二个控制点
        alphaPath.bezierCurveTo(rectX+400,rectY-50,rectX+50,rectY-100,rectX+180,rectY-200)
        alphaPath.bezierCurveTo(rectX,rectY-300,rectX-100,rectY-100,rectX-100,rectY-150)
        alphaPath.bezierCurveTo(rectX-200,rectY-200,rectX-250,rectY-100,rectX-250,rectY-50)
        alphaPath.bezierCurveTo(rectX-200,rectY+20,rectX-300,rectY+60,rectX-200,rectY+100)
        alphaPath.bezierCurveTo(rectX-150,rectY+160,rectX-50,rectY+100,rectX,rectY+150)
        alphaPath.bezierCurveTo(rectX+50,rectY+200,rectX+100,rectY+ 60,rectX+200,rectY+120)
        alphaPath.bezierCurveTo(rectX+250,rectY+160,rectX+100,rectY,rectX+250,rectY)
        alphaPath.closePath()
        this.shapePath = alphaPath
        this.draw(alphaPath)  
    }

    
    draw = (alphaPath:Path2D) =>{
        // 清空canvase
        let radius = this.dymRadius > 100 ? this.dymRadius-=0.3:this.dymRadius=200;
        this.bgCtx.clearRect(0,0,width,height);
        // this.bgCtx.beginPath()
        // this.bgCtx.strokeStyle="#fff"
        // // 第一段 起始点
        // this.bgCtx.moveTo(rectX+250,rectY)
        // // 第一个控制点 第二个控制点
        // this.bgCtx.bezierCurveTo(rectX+400,rectY-50,rectX+50,rectY-100,rectX+180,rectY-200)
        // this.bgCtx.bezierCurveTo(rectX,rectY-300,rectX-100,rectY-100,rectX-100,rectY-150)
        // this.bgCtx.bezierCurveTo(rectX-200,rectY-200,rectX-250,rectY-100,rectX-250,rectY-50)
        // this.bgCtx.bezierCurveTo(rectX-200,rectY+20,rectX-300,rectY+60,rectX-200,rectY+100)
        // this.bgCtx.bezierCurveTo(rectX-150,rectY+160,rectX-50,rectY+100,rectX,rectY+150)
        // this.bgCtx.bezierCurveTo(rectX+50,rectY+200,rectX+100,rectY+ 60,rectX+200,rectY+120)
        // this.bgCtx.bezierCurveTo(rectX+250,rectY+160,rectX+100,rectY,rectX+250,rectY)
        // this.bgCtx.closePath()
        // this.bgCtx.stroke()
        //  1000 点生成
        // this.bgCtx.clearRect(0,0,width,height)
        // 每帧生成 1000个点
        // 已经生成的点保存 后动态运动 明天写一下
        for (let index = 0; index < 30; index++) {
            // 生成60 - 120之间的随机数字
            let x = -Math.floor(Math.random() * radius);
            // 求对应点的y值
            let y = Math.sqrt(Math.pow(radius,2) - Math.pow(x,2));
            let newX = Math.random() < 0.5 ? x : -x;
            let newY = Math.random() < 0.5 ? y : -y;
            let particle = new Particle(newX+rectX,newY+rectY);
            this.particles.push(particle);
        }

        // // 生成60 - 120之间的随机数字
        // let x = -Math.floor(Math.random() * radius);
        // // 求对应点的y值
        // let y = Math.sqrt(Math.pow(radius,2) - Math.pow(x,2));
        // let newX = Math.random() < 0.5 ? x : -x;
        // let newY = Math.random() < 0.5 ? y : -y;
        // let particle = new Particle(newX+rectX,newY+rectY);
        // this.particles.push(particle);
        
        for (let index = 0; index < this.particles.length; index++) {
            const particle = this.particles[index];
            let isAlpha = this.bgCtx.isPointInPath(alphaPath,particle.x,particle.y,"nonzero")
            // 此处为圆形Remove 该处需要被重写
            let isRemove = Math.pow(Math.abs(particle.x - rectX),2) + Math.pow(Math.abs(particle.y- rectY),2) - Math.pow(300,2);
            if (isRemove>0) {
                // console.log(index,this.particles.length,isAlpha)
                this.particles.splice(index,1);
            }
            particle.render(this.bgCtx,isAlpha);
            
        }

        this.bgCtx.beginPath();
        this.bgCtx.fillStyle = "#363636";
        this.bgCtx.arc(rectX,rectY,radius,0,Math.PI*2,true);
        this.bgCtx.fill();

        window.requestAnimationFrame(function(){
            this.draw(alphaPath)
        }.bind(this));
        // setInterval(this.draw,5000)
    }

    render(){
        return (
            <div>
                <canvas width={width} height={height} className="bgCanvas" ref={this.bgCanvas} />
            </div>
        )
    }
}