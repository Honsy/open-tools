import Home from './pages/home'
import Test from './pages/test'
import Calculator from './pages/calculator'
import Crypto from './pages/crypto'
// import Prettier from './pages/prettier'
import RGB from './pages/rgb'
import HexConvert from './pages/hexConvert'
import Protobuf from './pages/protobuf'
import Moment from './pages/moment'
import PDF from './pages/pdf'
import Xterm from './pages/xterm'
import Canvas from './pages/canvas'
import Translate from './pages/translate'

export default [
    {
        path: "/",
        exact: true,
        component: Home
    },{
        path: "/test",
        exact: true,
        component: Test
    },{
        path: "/calculator",
        exact: true,
        component: Calculator
    },{
        path: "/crypto",
        exact: true,
        component: Crypto
    },{
        path: "/rgb",
        exact: true,
        component: RGB
    },{
        path: "/hexconvert",
        exact: true,
        component: HexConvert
    },{
        path: "/protobuf",
        exact: true,
        component: Protobuf
    },{
        path: "/moment",
        exact: true,
        component: Moment
    },{
        path: "/pdf",
        exact: true,
        component: PDF
    },{
        path: "/xterm",
        exact: true,
        component: Xterm
    },{
        path: "/canvas",
        exact: true,
        component: Canvas
    },{
        path: "/translate",
        exact: true,
        component: Translate
    }
]