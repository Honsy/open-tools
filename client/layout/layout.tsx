import * as React from 'react'
import Header from './header'
import Footer from './footer'
import Slide from './sidebar'
import config from "./../config";


function Layout(props){
    const { children } = props
    return (
        <div className="layout">
            <Header></Header>
            <div className="base">
                <Slide className="sidebar" data={config.tools}></Slide>
                <div className="base-container">
                    <div className="children">
                        {children}
                    </div>
                    <Footer></Footer>
                </div>
            </div>
        </div>
    )
}

export default Layout