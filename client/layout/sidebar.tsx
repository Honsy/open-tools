import * as React from 'react'
import { ToolObj } from './../interface'

interface SideBarProps{
    data:ToolObj[]
    className?:string
}

function SideBar(props:SideBarProps){
    const { data } = props


    return (
        <div className={props.className}>
            {
                data.map((item,index)=>{
                    return (
                        <div key={"side"+index}>
                            {item.title}
                        </div>
                    )
                })
            }

        </div>
    )
}

export default SideBar