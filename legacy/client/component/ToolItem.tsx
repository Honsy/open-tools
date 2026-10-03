import * as React from 'react'
import {Link} from 'react-router-dom'

interface ToolItemProps {
    title:string
    href:string
    color?:string
    info?:string
}

function ToolItem(props:ToolItemProps){
    const { title,href,color,info } = props
    return (
        <div className="toolitem" style={{backgroundColor:color?color:"#555"}}>
            <a href={href}>
                <p className="title">{title}</p>
                <p className="info">{info?info:""}</p>
            </a>
        </div>
    )
}

export default ToolItem