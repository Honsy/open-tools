import React from 'react'
import {Layout} from './../layout'

interface RMBConvertState{
    waitConvertRMB:string;
    waitConvertLetter:string;
    rmbResult:string;
    letterResult:string;
}

export default class RMBConvert extends React.Component<{},RMBConvertState>{
    
    static title = "常用文字转换-OB工具"

    constructor(props){
        super(props)
        this.state = {
            waitConvertRMB:"",
            waitConvertLetter:"",
            rmbResult:"",
            letterResult:""
        }
    }

    // 人民币大写转换
    rmbNumberToChinese = () =>{
        let { waitConvertRMB } = this.state
        let chineses = "零壹贰叁肆伍陆柒捌玖"
        let units = "京亿万仟佰拾兆万仟佰拾亿仟佰拾万仟佰拾元角分"
        let result = ""
        if (waitConvertRMB.indexOf(".") > 0) {
            let splitArray = waitConvertRMB.split(".")
            waitConvertRMB = splitArray[1].length > 1?waitConvertRMB.replace(".",""):waitConvertRMB.replace(".","")+"0"
        }else{
            waitConvertRMB += "00"
        }
        for (let index = 0; index < waitConvertRMB.length; index++) {
            const element = waitConvertRMB[index];
            let unit = units[units.length - waitConvertRMB.length + index]
            if (unit == "角" || unit == "分") {
                result += Number(element)===0?"":chineses.charAt(Number(element)) + units[units.length - waitConvertRMB.length + index]

            }else{
                result += chineses.charAt(Number(element)) + units[units.length - waitConvertRMB.length + index]
            }
        }
        result += "整"
        this.setState({
            rmbResult:result
        })
    }

    // 英文大小写转换
    captialToLowerCase = () =>{
        const { waitConvertLetter } = this.state

    }

    inputValueChange = (name:keyof RMBConvertState,value:any) =>{
        this.setState({
            [name]:value
        } as Pick<RMBConvertState,keyof RMBConvertState>)
    }

    render(){
        const { waitConvertRMB,rmbResult,waitConvertLetter } = this.state
        return (
            <Layout>
                <div className="container rmbconvert">
                    <section className="px-5 py-5">
                        <h4 className="title">人民币大写转换</h4>
                        <div className="field has-addons">
                            <p className="control">
                                <input className="input" value={waitConvertRMB} type="text" placeholder="输入待转换的金额" onChange={(e)=>this.inputValueChange("waitConvertRMB",e.target.value)}/>
                            </p>
                            <p className="control">
                            <button className="button" onClick={this.rmbNumberToChinese}>
                                转换
                            </button>
                            </p>
                        </div>
                        <p>{rmbResult}</p>
                    </section>
                    <section className="px-5 py-5">
                        <h4 className="title">英文字母大小写转换</h4>
                        <div className="columns letter">
                            <div className="column field has-addons is-two-thirds lt-input">
                                <p className="control">
                                    <input className="input" value={waitConvertLetter} type="text" placeholder="输入待转换的英文字母" onChange={(e)=>this.inputValueChange("waitConvertLetter",e.target.value)}/>
                                </p>
                                <p className="control">
                                <button className="button" onClick={this.captialToLowerCase}>
                                    转换
                                </button>
                                </p>
                            </div>
                            <div className="column">
                                <div className="control">
                                    <label className="radio">
                                        <input type="radio" name="answer" />
                                        大写
                                    </label>
                                    <label className="radio">
                                        <input type="radio" name="answer" />
                                        小写
                                    </label>
                                </div>
                            </div>
                        </div>
                        <p></p>
                    </section>
                </div>
            </Layout>
        )
    }
}