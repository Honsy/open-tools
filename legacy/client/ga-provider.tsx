
import * as React from 'react'

let ga = {}

const context = React.createContext(ga);


export default function GAProvider(props: {
    children: any,
    data:any
}) {
    return <context.Provider value={props.data}>{props.children}</context.Provider>;
}

GAProvider.context = context