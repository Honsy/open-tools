import * as React from 'react'
import {renderToString} from 'react-dom/server'
import {StaticRouter} from 'react-router-dom'
import App from './app'

interface ComponentProps {
    url:string
}

interface SSRProps {
    componentName:string,
    data:ComponentProps
}

export default function ssr(data:SSRProps){
    var html
    const context = {};
    switch (data.componentName) {
        case 'SPA':
            html =  renderToString(
                <StaticRouter location={data.data.url} context={context}>
                    <App/>
                </StaticRouter>
            )
            break;
    
        default:
            break;
    }

    return {html:html}
}
