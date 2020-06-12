import React,{ useEffect, useState } from 'react';

export interface RC4Config{
    password?:string
}

interface RC4ConfigProps{
    cfg?:RC4Config
    onResult?:(cfg:RC4Config)=>void
}

function RC4CryptoConfig(props:RC4ConfigProps) {
    const { cfg,onResult } = props
    const [config,setConfig] = useState(cfg)

    function inputValueChange(value:any){
        config.password = value
        setConfig(config)
        onResult?onResult(config):null
    }

    return (
        <div className="columns cry-cfg">
            <div className="column">
                <span>密码</span>
                <div>
                <input className="input" type="text" onChange={(e)=>inputValueChange(e.target.value)}/>
                </div>
            </div>
        </div>
    )
}

export default RC4CryptoConfig
