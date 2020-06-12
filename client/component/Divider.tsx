import * as React from 'react'

interface DividerProps{
    type?: 'horizontal' | 'vertical';
}

function Divider(props:DividerProps){
    const { type } = props

    return (
        <div className={"divider-"+type}>

        </div>
    )
}

export default Divider