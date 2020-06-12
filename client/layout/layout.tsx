import * as React from 'react'
import Header from './header'
import Footer from './footer'
import Slide from './slide'

function Layout(props){
    const { children } = props
    return (
        <div>
            <Header></Header>
            {children}
            <Footer></Footer>
        </div>
    )
}

export default Layout