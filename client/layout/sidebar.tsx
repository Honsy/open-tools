import React,{useState} from 'react'
import { ToolObj } from './../interface'
import {useHistory} from "react-router-dom"
// import config from './../../config'

interface SideBarProps{
    data:ToolObj[]
    className?:string
}

function SideBar(props:SideBarProps){
    let history = useHistory()
    let nowSelectIndex = "0-0"
    const { data } = props
    
    for (let index = 0; index < data.length; index++) {
        let element = data[index];
        for (let jndex = 0; jndex < element.data.length; jndex++) {
            let selement =  element.data[jndex];
            if (selement.href == history.location.pathname) {
                nowSelectIndex = index+"-"+jndex
            }
        }
    }

    const [selectIndex,setSelectIndex] = useState(nowSelectIndex)


    return (
        <div className={props.className}>
            <aside className="menu mx-2 mt-2">
            {
                data.map((item,topIndex)=>{
                    return (
                        <React.Fragment key={"side"+topIndex}>
                            <p className="menu-label">{item.title}</p>
                            <ul className="menu-list">
                                {
                                    item.data.map((tool,index)=>{
                                        return (
                                            <li key={tool.title} onClick={()=>setSelectIndex(topIndex+"-"+index)}>
                                                <a className={selectIndex===(topIndex+"-"+index)?"is-active":""} href={tool.href}>{tool.title}</a>
                                            </li>
                                        )
                                    })
                                }
                            </ul>
                        </React.Fragment>
                    )
                })
            }
            </aside>
        </div>
    )
}

export default SideBar