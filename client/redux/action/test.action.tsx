import constant from './../constant'

export function testAction(data){
    return {type:constant.TEST,data}
}

export function testActionSuccess(data){
    return {type:constant.TEST_SUCCESS,data}
}
export function testActionError(data){
    return {type:constant.TEST_ERROR,data}
}
