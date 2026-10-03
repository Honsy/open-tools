import constant from './../constant'

export function testReducer(state:any = {},action:any){
    switch (action.type){
        case constant.TEST_SUCCESS:
            return {
                ...state,
            }
        case constant.TEST_ERROR:
            return {
            }
        default:
            return state
    }
}