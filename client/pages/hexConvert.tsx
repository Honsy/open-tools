import * as React from "react";
import { Layout } from "./../layout";
import GAProvider from "./../ga-provider";

interface HexItem {
    title:string
    type:number
}

interface HexConvertState {
    resultValue:string | number
    selectIndex:number
    resultSelectIndex:number
    inputValue:string
}

export default class HexConvert extends React.Component<{},HexConvertState>{
  static contextType = GAProvider.context;

  HexArray: HexItem[] = [
    { title: "2进制", type: 2 },
    { title: "4进制", type: 4 },
    { title: "8进制", type: 8 },
    { title: "10进制", type: 10 },
    { title: "16进制", type: 16 },
    { title: "32进制", type: 32 },
  ];

  static initData() {
    let hexConver = new HexConvert(null);
    return {
      hexArray: hexConver.HexArray,
    };
  }

  constructor(props: any) {
    super(props);
    this.state = {
        resultValue:"",
        selectIndex:4,
        resultSelectIndex:5,
        inputValue:""
    }
  }

  onSelectIndexChange = (type:number,value:number) =>{
    if (type == 0) {
        this.setState({
            selectIndex:value
        })
    }else{
        this.setState({
            resultSelectIndex:value
        },()=>this.onHexChange(true,String(this.state.inputValue)))
    }
  }

  onHexChange = (isResult:boolean,value:string):void =>{
    const { selectIndex,resultSelectIndex } = this.state
    let selectItem = this.HexArray[selectIndex]
    let resultSelectItem = this.HexArray[resultSelectIndex]
    let result = null
    // 是否为结果进制变换
    if (value.indexOf(".")>-1) {
        result = Number(value).toString(selectItem.type)
    }else{
        result = parseInt(value,selectItem.type)
    }
    let output = result.toString(resultSelectItem.type)

    this.setState({
        resultValue:output,
        inputValue:value
    })
  }

  render() {
    const { hexArray } = this.context;
    const { resultValue,selectIndex,resultSelectIndex,inputValue } = this.state
    return (
      <Layout>
        <div className="container hexconvert">
          <h4>进制转换</h4>
          <div className="start is-spaced">
            <div className="control">
                {
                    hexArray.map((item, index) => {
                        return (
                            <label className="radio" key={item.title+"asdsa"}>
                                <input type="radio" key={item.title+"start"} name="start" checked={selectIndex===index} onChange={()=>this.onSelectIndexChange(0,index)}/>
                                {item.title}
                            </label>
                        );
                    })
                }
            </div>
            <div className="field">
                <p className="control">
                    <input className="input" value={inputValue} placeholder="转换数字" onChange={(e)=>this.onHexChange(false,e.target.value)}/>
                </p>
            </div>
          </div>
          <div className="result mt-4 mb-4">
            <div className="control">
                {
                    hexArray.map((item, index) => {
                        return (
                            <label className="radio" key={item.title+"result"}>
                                <input type="radio" key={item.title+"in"} name="result" checked={resultSelectIndex===index} onChange={()=>this.onSelectIndexChange(1,index)}/>
                                {item.title}
                            </label>
                        );
                    })
                }
            </div>
            <div className="field">
                <p className="control">
                <input className="input" placeholder="转换结果" value={resultValue} readOnly/>
                </p>
            </div>
          </div>
        </div>
      </Layout>
    );
  }
}
