import * as React from "react";

const hashArrays = [
    "md5","sha1","sha256","sha224","sha512","sha384","sha3","ripemd160"
]

function HashCryptoConfig(props) {

    return (
        <div className="control">
            {
                hashArrays.map((item,_)=>{
                    return (
                        <label key={item} className="radio">
                            <input type="radio" name="foobar"/>
                            &nbsp;{item}
                        </label>
                    )
                })
            }
        </div>
    )
}

export default HashCryptoConfig
