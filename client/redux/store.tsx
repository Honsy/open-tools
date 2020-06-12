import {createStore,applyMiddleware,Store} from 'redux'
import reducer from './reducers'
import createSagaMiddleware from 'redux-saga'; // 1：saga 引入createSagaMiddleware
import { composeWithDevTools } from 'redux-devtools-extension'; // 调试工具
// import rootSaga from './sagas'

const sagaMiddleware = createSagaMiddleware()

const middleware = applyMiddleware(
    sagaMiddleware, 
)
// 声明store
let clientStore:any,serverStore:any;

export function createClientStore(preloadState:any):Store{
    if (clientStore) {
        return clientStore
    }
    clientStore = createStore(
        reducer,
        preloadState,
        composeWithDevTools(
            middleware
        )
    )
    // 异步
    // sagaMiddleware.run(rootSaga)
    return clientStore
}

export function createServerStore(preloadState:any){
    serverStore = createStore(
        reducer,
        preloadState,
        middleware
    )
    // 异步
    // sagaMiddleware.run(rootSaga)
    return serverStore
}
