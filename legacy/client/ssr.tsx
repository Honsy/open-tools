import * as React from 'react'
import {renderToString} from 'react-dom/server'
import {StaticRouter} from 'react-router-dom'
import App from './app'
import Home from './pages/home/home'
import Prettier from './pages/prettier/prettier'

import GAProvider from './ga-provider'
import { Provider } from 'react-redux';

interface ComponentProps {
    url:string,
    data:object
}

interface SSRProps {
    componentName:string,
    data:ComponentProps
}

export default function ssr(data:SSRProps,store:any){
    var html
    const context = {};
    switch (data.componentName) {
        case 'SPA':
            html =  renderToString(
                <StaticRouter location={data.data.url}>
                    <GAProvider data={data.data.data}>
                        <Provider store={store}>
                            <App/>
                        </Provider>
                    </GAProvider>
                </StaticRouter>
            )
            break;
        case 'home':
            html =  renderToString(
                <StaticRouter location={data.data.url}>
                    <GAProvider data={data.data.data}>
                        <Provider store={store}>
                            <Home/>
                        </Provider>
                    </GAProvider>
                </StaticRouter>
            )
            break;
        case 'prettier':
            html =  renderToString(
                <StaticRouter location={data.data.url}>
                    <GAProvider data={data.data.data}>
                        <Provider store={store}>
                            <Prettier/>
                        </Provider>
                    </GAProvider>
                </StaticRouter>
            )
            break;
        default:
            break;
    }

    return {html:html}
}
