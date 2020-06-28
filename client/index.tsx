import "@babel/polyfill"
// import "core-js/stable";
// import "regenerator-runtime/runtime";
import * as React from 'react'
import * as ReactDOM from 'react-dom'
import App from './app'
import {BrowserRouter} from 'react-router-dom'
import GAProvider from './ga-provider'
import { Provider } from 'react-redux'
import { createClientStore }  from './redux/store'

let preloadState = window.STATE_FROM_SERVER
// 创建store
const store = createClientStore(preloadState)

let container = document.getElementById('react-container');
if (container) {
    let APP_DATA = {}
    let jsData = document.getElementById("js-data")
    if (jsData) {
        APP_DATA = jsData.textContent !== "undefined"?JSON.parse(jsData.textContent):{}
    }
    ReactDOM.hydrate(
        <BrowserRouter>
            <GAProvider data={APP_DATA}>
                <Provider store={store}>
                    <App/>
                </Provider>
            </GAProvider>
        </BrowserRouter>
        ,document.getElementById('react-container')
    )
}