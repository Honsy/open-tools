// import * as React from "react";
import React,{ useEffect, useState } from 'react';

export interface AESConfig{
  mode?:"ECB"|"CBC"|"CTR"|"OFB"|"CFB"
  padding?:"Pkcs7"|"Iso97971"|"AnsiX923"|"Iso10126"|"ZeroPadding"|"NoPadding"
  password?:string
  iv?:string
  outputType?:"BASE64"|"HEX"
}

interface AESConfigProps{
  cfg?:AESConfig
  onResult?:(cfg:AESConfig)=>void
}

function AESCryptoConfig(props:AESConfigProps) {
  const { cfg,onResult } = props

  const [config,setConfig] = useState(cfg)


  function inputValueChange(key:string,value:any){
    if (key==="password") {
      config.password = value
    }else{
      config.iv = value
    }
    setConfig(config)
    onResult?onResult(config):null
  }

  return (
    <div className="columns cry-cfg">
      <div className="column mode">
        <span>加密模式</span>
        <div className="select">
          <select defaultValue={config.mode}>
            <option value="ECB">ECB</option>
            <option value="CBC">CBC</option>
            <option value="CTR">CTR</option>
            <option value="OFB">OFB</option>
            <option value="CFB">CFB</option>
          </select>
        </div>
      </div>
      <div className="column">
        <span>填充</span>
        <div className="select">
          <select defaultValue={config.padding}>
            <option value="NoPadding">NoPadding</option>
            <option value="Pkcs7">Pkcs7</option>
            <option value="Iso97971">Iso97971</option>
            <option value="AnsiX923">AnsiX923</option>
            <option value="Iso10126">Iso10126</option>
            <option value="ZeroPadding">ZeroPadding</option>
          </select>
        </div>
      </div>
      <div className="column">
        <span>密码</span>
        <div>
          <input value={config.password} className="input" type="text" onChange={(e)=>inputValueChange("password",e.target.value)}/>
        </div>
      </div>
      <div className="column">
        <span>偏移量</span>
        <div>
          <input value={config.iv} className="input" type="text" onChange={(e)=>inputValueChange("iv",e.target.value)}/>
        </div>
      </div>
      {/* <div className="column">
        <span>输出</span>
        <div className="select">
          <select>
            <option>BASE64</option>
            <option>HEX</option>
          </select>
        </div>
      </div>
      <div className="column">
        <span>字符集</span>
        <div className="select">
          <select>
            <option>GB2312</option>
            <option>GBK</option>
            <option>UTF8</option>
          </select>
        </div>
      </div> */}
    </div>
  );
}

export default AESCryptoConfig;
