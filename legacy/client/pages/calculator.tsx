import * as React from 'react'
import {Layout} from './../layout'

interface CalculatorState{
    inputValue:string
}

export default class Calculator extends React.Component<{},CalculatorState>{

    static title = "计算器-OB工具";

    static initData(){

    }

    btnlist:any

    constructor(props){
        super(props)
        this.state = {
            inputValue:"0"
        }
    }

    componentDidMount(){

    }

    // 运算符点击
    operatorClick = (type) =>{
        
    }

    // 小数点点击
    dianClick = () =>{
        const { inputValue } = this.state
        if (inputValue.indexOf(".") > -1) {
            return null;
        }
        let val = inputValue + "."
        this.setState({
            inputValue:val
        })
    }

    // 数字点击
    numClick = (num) =>{
        const { inputValue } = this.state
        let val = inputValue
        if (String(num) === "0" && inputValue==="0") {
            return null;
        }else if(inputValue ==="0" && String(num) !== "0"){
            val = String(num)
        }else{
            val = inputValue + String(num)
        }

        this.setState({
            inputValue:val
        })
    }

    render(){
        const { inputValue } = this.state
        return (
            <Layout>
                <div className="container calculator">
                    <div className="top">

                    </div>
                    <div className="console">
                        <div className="last-input">

                        </div>
                        <div className="input">
                            {inputValue}
                        </div>
                    </div>
                    <div className="columns list">   
                        <div className="column is-one-quarter other">%</div>
                        <div className="column is-one-quarter other">CE</div>
                        <div className="column is-one-quarter other">C</div>
                        <div className="column is-one-quarter other">撤销</div>
                    </div>
                    <div className="columns list">   
                        <div className="column is-one-quarter num" onClick={()=>this.numClick(7)}>7</div>
                        <div className="column is-one-quarter num" onClick={()=>this.numClick(8)}>8</div>
                        <div className="column is-one-quarter num" onClick={()=>this.numClick(9)}>9</div>
                        <div className="column is-one-quarter other">X</div>
                    </div>
                    <div className="columns list">   
                        <div className="column is-one-quarter num" onClick={()=>this.numClick(4)}>4</div>
                        <div className="column is-one-quarter num" onClick={()=>this.numClick(5)}>5</div>
                        <div className="column is-one-quarter num" onClick={()=>this.numClick(6)}>6</div>
                        <div className="column is-one-quarter other" onClick={()=>this.operatorClick(7)}>-</div>
                    </div>
                    <div className="columns list">   
                        <div className="column is-one-quarter num" onClick={()=>this.numClick(1)}>1</div>
                        <div className="column is-one-quarter num" onClick={()=>this.numClick(2)}>2</div>
                        <div className="column is-one-quarter num" onClick={()=>this.numClick(3)}>3</div>
                        <div className="column is-one-quarter other">+</div>
                    </div>
                    <div className="columns list">   
                        <div className="column is-one-quarter num">+/-</div>
                        <div className="column is-one-quarter num" onClick={()=>this.numClick(0)}>0</div>
                        <div className="column is-one-quarter num" onClick={()=>this.dianClick()}>.</div>
                        <div className="column is-one-quarter other">=</div>
                    </div>
                </div>
            </Layout>
        )
    }
}