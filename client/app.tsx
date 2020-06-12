import * as React from 'react'
import { Switch,Route} from 'react-router-dom'
import routerConfig from './router.config'

class App extends React.Component{
    constructor(props:any){
        super(props)
    }

    render(){
        return (
            <Switch>
                {
                    routerConfig.map((item,index)=>{
                        return (
                            <Route path={item.path} key={index} exact={item.exact} component={item.component}></Route>
                        )
                    })
                }
            </Switch>
        )
    }
}

export default App