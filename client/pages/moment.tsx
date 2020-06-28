import * as React from 'react'
import {Layout} from './../layout'
import moment from 'moment'

enum SecondType{
    ms = 0,
    s = 1
}

interface MomentState{
    oneSecondType: SecondType;
    twoSecondType: SecondType;
    oneTimeStamp:string;
    oneTimeUTC:string;
    twoTimeUTC:string;
    twoTimeStamp:string;
}

const timeFormat:string = "YYYY-MM-DD hh:mm:ss"

export default class Moment extends React.Component<{},MomentState>{
    constructor(props:any){
        super(props)
        this.state = {
            oneSecondType:SecondType.ms,
            oneTimeStamp:String(moment().valueOf()),
            oneTimeUTC:"",
            twoSecondType:SecondType.ms,
            twoTimeUTC:moment().format(timeFormat),
            twoTimeStamp:""
        }
    }

    static title = "时间转换-OB工具-时间戳转标准时间,Moment转换"

    componentDidMount(){
        this.timeConvert(0)
        this.timeConvert(1)
    }

    timeConvert = (type:number) =>{
        const { oneSecondType,oneTimeStamp,oneTimeUTC,twoSecondType,twoTimeStamp,twoTimeUTC }  = this.state

        if (type === 0) {
            let dayMoment = Number(oneSecondType) === SecondType.ms ? moment(Number(oneTimeStamp)) : moment.unix(Number(oneTimeStamp))
            let format = dayMoment.format(timeFormat)
            this.setState({
                oneTimeUTC:format
            })
        }else{
            let timestamp = Number(twoSecondType) === SecondType.ms ? moment(twoTimeUTC).valueOf() : moment(twoTimeUTC).unix()
            this.setState({
                twoTimeStamp:String(timestamp)
            })
        }
    }

    inputValueChange = (name:keyof MomentState,value:string | SecondType,e) =>{
        this.setState({
            [name]:value
        } as Pick<MomentState,keyof MomentState>)
    }

    render(){
        const { oneSecondType,oneTimeStamp,oneTimeUTC,twoSecondType,twoTimeStamp,twoTimeUTC } = this.state
        return (
            <Layout>
                <div className="container moment">
                    <div className="time-convert mb5">
                        <span className="form-title">时间戳：</span>
                        <input className="input" type="text" value={oneTimeStamp} onChange={(e)=>this.inputValueChange("oneTimeStamp",e.target.value,e)}/>
                        <div className="select">
                            <select value={oneSecondType} onChange={(e)=>this.inputValueChange("oneSecondType",e.target.value,e)}>
                                <option value={SecondType.ms}>毫秒(ms)</option>
                                <option value={SecondType.s}>秒(s)</option>
                            </select>
                        </div>
                        <button className="button convert is-primary is-light" onClick = {()=>this.timeConvert(0)}>转换</button>
                        <input className="input" type="text" value={oneTimeUTC} readOnly/>
                        <span>北京时间</span>
                    </div>
                    <div className="time-convert">
                        <span className="form-title">时间：</span>
                        <input className="input" type="text" value={twoTimeUTC}  onChange={(e)=>this.inputValueChange("twoTimeUTC",e.target.value,e)}/>
                        <span>北京时间</span>
                        <button className="button convert is-primary is-light"  onClick = {()=>this.timeConvert(1)}>转换</button>
                        <input className="input" type="text" readOnly value={twoTimeStamp}/>
                        <div className="select">
                            <select value={twoSecondType} onChange={(e)=>this.inputValueChange("twoSecondType",e.target.value,e)}>
                                <option value={SecondType.ms}>毫秒(ms)</option>
                                <option value={SecondType.s}>秒(s)</option>
                            </select>
                        </div>
                    </div>
                </div>
            </Layout>
        )
    }
}