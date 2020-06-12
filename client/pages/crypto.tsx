import * as React from "react";
import { Layout } from "./../layout";
import { Divider,AESCryptoConfig,RC4CryptoConfig,HashCryptoConfig,AESConfig,RC4Config } from './../component'
import GAProvider from "./../ga-provider";
import sha256 from "crypto-js/sha256";
import MD5 from "crypto-js/md5";
import AES from "crypto-js/aes";
import RC4 from "crypto-js/rc4";
import Base64 from 'crypto-js/enc-base64';
import HEX from 'crypto-js/enc-hex';
import Utf8 from 'crypto-js/enc-utf8';
import CFB from 'crypto-js/mode-cfb'
import CTR from 'crypto-js/mode-ctr'
import CTRGladman from 'crypto-js/mode-ctr-gladman'
import OFB from 'crypto-js/mode-ofb'
import ECB from 'crypto-js/mode-ecb'
import pkcs7 from 'crypto-js/pad-pkcs7'
import ansix923 from 'crypto-js/pad-ansix923'
import iso10126 from 'crypto-js/pad-iso10126'
import iso97971 from 'crypto-js/pad-iso97971'
import zeropadding from 'crypto-js/pad-zeropadding'
import nopadding from 'crypto-js/pad-nopadding'

interface TypeItem {
  title:string
  type:number
}

interface CryptoState {
  decryptText:string
  encryptText:string
  selectIndex:number
}

export default class Crypto extends React.Component<{},CryptoState> {
  static contextType = GAProvider.context;

  static title = "加密解密-BASE64,AES,MD5,DC4....";
  // 数据获取
  static async initData(params: any) {
    let list:TypeItem[] = [{ title: "BASE64加密解密", type: 0 },
    { title: "AES加密解密", type: 1 },
    { title: "MD5加密解密", type: 2 },
    { title: "RC4加密解密", type: 3 },
    { title: "Rabbit加密解密", type: 4 },
    { title: "evpkdf加密解密", type: 5 },
    { title: "tripledes加密解密", type: 6 },
    { title: "哈希加密解密", type: 7 }];
    return { content: "首页", list: list };
  }

  encryptTextArea:React.RefObject<HTMLTextAreaElement>
  decryptTextArea:React.RefObject<HTMLTextAreaElement>
  aesCfg:AESConfig
  rc4Cfg:RC4Config


  constructor(props: any) {
    super(props)
    this.encryptTextArea = React.createRef()
    this.decryptTextArea = React.createRef()
    this.aesCfg = {
      mode:"CBC",
      padding:"Pkcs7",
      outputType:"BASE64"
    }
    this.rc4Cfg = {

    }
    this.state = {
      decryptText:"",
      encryptText:"",
      selectIndex:0,
    }
  }

  componentDidMount() {

  }
  // 类型变化
  typeChange = (index:number) =>{
    this.setState({
      selectIndex:index
    })
  }
  // 解密
  decrypt = () =>{
    const { list } = this.context;
    const { selectIndex,decryptText } = this.state;
    let selectItem = list[selectIndex]

    switch (selectItem.type) {
      case 0:
        this.setState({
          encryptText:this.base64(false,decryptText)
        },()=>{
          this.encryptTextArea.current.style.height = this.encryptTextArea.current.scrollHeight + "px"
        })
        break;
      case 1:
        this.setState({
          encryptText:this.aes(false,decryptText)
        },()=>{
          this.encryptTextArea.current.style.height = this.encryptTextArea.current.scrollHeight + "px"
        })
        break; 
      case 3:
        this.setState({
          encryptText:this.rc4(false,decryptText)
        },()=>{
          this.encryptTextArea.current.style.height = this.encryptTextArea.current.scrollHeight + "px"
        })
        break;    
      default:
        break;
    }
  }

  // 加密
  encrypt = () =>{
    const { list } = this.context;
    const { selectIndex,encryptText } = this.state;
    let selectItem = list[selectIndex]

    switch (selectItem.type) {
        // BASE64
      case 0:
        this.setState({
          decryptText:this.base64(true,encryptText)
        },()=>{
          this.decryptTextArea.current.style.height = this.decryptTextArea.current.scrollHeight + "px"
        })
        break;
        // AES
      case 1:
        this.setState({
          decryptText:this.aes(true,encryptText)
        },()=>{
          this.decryptTextArea.current.style.height = this.decryptTextArea.current.scrollHeight + "px"
        })
        break;
        // MD5
      case 2:
        this.setState({
          decryptText:this.md5(encryptText)
        },()=>{
          this.decryptTextArea.current.style.height = this.decryptTextArea.current.scrollHeight + "px"
        })
        break;
        // RC4
      case 3:
        this.setState({
          decryptText:this.rc4(true,encryptText)
        },()=>{
          this.decryptTextArea.current.style.height = this.decryptTextArea.current.scrollHeight + "px"
        })
        break;
      default:
        break;
    }
  }

  // BASE64加解密
  base64 = (en:boolean,words:string):string =>{
    return en?Base64.stringify(Utf8.parse(words)):Utf8.stringify(Base64.parse(words))
  }
  // AES加解密
  aes = (en:boolean,words:string):string =>{
    let result = ""
    let password = this.aesCfg.password?this.aesCfg.password:""
    let pad = this.aesCfg.padding
    let config:any = {}
    // mode 找不到CBC的mode在哪 默认为CBCmode
    if (this.aesCfg.mode !== "CBC") {
      config.mode = this.aesCfg.mode==="CFB"?CFB:this.aesCfg.mode==="CTR"?CTR:this.aesCfg.mode==="ECB"?ECB:this.aesCfg.mode==="OFB"?OFB:null
    }
    // padding
    config.padding = pad==="AnsiX923"?ansix923:pad==="Iso10126"?iso10126:pad==="Iso97971"?iso97971:pad==="NoPadding"?nopadding:pad==="Pkcs7"?pkcs7:pad==="ZeroPadding"?zeropadding:null
    // iv
    if (this.aesCfg.iv) {
      config.iv = this.aesCfg.iv
    }
    if (en) {
      let sub= AES.encrypt(words,password,config);
      result = sub.toString()
    }else{
      let sub= AES.decrypt(words,password,config);
      result =sub.toString(Utf8)
    }
    return result
  }
  // md5加密
  md5 = (words:string):string =>{
    return MD5(words)
  }
  // rc4
  rc4 = (en:boolean,word:string):string =>{
    let password = this.rc4Cfg.password?this.rc4Cfg.password:""
    return en?RC4.encrypt(word,password):RC4.decrypt(word,password).toString(Utf8)
  }


  // UI重置
  resize = () =>{

  }

  // 数据接收
  inputValueChange = (key: keyof CryptoState,value:string|number) =>{
    this.setState({
      [key]:value
    } as Pick<CryptoState,keyof CryptoState>)
  }

  // 子组件数据接收
  aesConfigChange = (cfg:AESConfig) =>{
    this.aesCfg = cfg
    console.log(cfg,this.aesCfg)
  }

  rc4ConfigChange = (cfg:AESConfig) =>{
    this.rc4Cfg = cfg
  }

  render() {
    const { list } = this.context;
    const { decryptText,encryptText,selectIndex} = this.state
    let selectItem = list[selectIndex]

    return (
      <Layout>
        <div className="container crypto">
          <div className="row types">
                {list.map((item, index) => {
                  return (
                    <a key={item.title} 
                      className={selectIndex === index?"button is-dark":"button"}
                      onClick={()=>this.typeChange(index)}
                      >
                      {item.title}
                    </a>
                  );
                })}
          </div>
          <div className="configure">
                {
                  // AES配置
                  selectItem.type === 1?
                    (<AESCryptoConfig cfg={this.aesCfg} onResult={this.aesConfigChange}></AESCryptoConfig>):
                    selectItem.type === 3?
                    (<RC4CryptoConfig cfg={this.rc4Cfg} onResult={this.rc4ConfigChange}></RC4CryptoConfig>):
                    selectItem.type === 7?
                    (<HashCryptoConfig></HashCryptoConfig>):null
                }
          </div>
          <div className="content">
            <div className="columns">
                <div className="column is-two-fifths">
                  <label className="active">待加密文本</label>
                  <textarea 
                      ref={this.encryptTextArea}
                      value={encryptText} 
                      style={{minHeight:200}} 
                      id="textarea1" 
                      className="textarea" 
                      onChange={(e)=>this.inputValueChange("encryptText",e.target.value)}></textarea>
                </div>
                <div className="column center">
                  <a className="button" onClick={this.encrypt}>加密</a>
                  <Divider type="horizontal"></Divider>
                  <button disabled={selectItem.type === 2} className="button" onClick={this.decrypt}>解密</button >
                </div>
                <div className="column is-two-fifths">
                  <label className="active">待解密文本</label>
                  <textarea
                      ref={this.decryptTextArea}
                      value={decryptText} 
                      style={{minHeight:200}} 
                      id="textarea1" 
                      className="textarea" 
                      onChange={(e)=>this.inputValueChange("decryptText",e.target.value)}></textarea>
                </div>
            </div>
          </div>
        </div>
      </Layout>
    );
  }
}
