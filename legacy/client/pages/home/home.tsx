import * as React from "react";
import GAProvider from "../../ga-provider";
import { Layout } from "../../layout";
import { ToolItem } from "../../component";
import BGCanvas2 from "../../component/canvas/bgCanvas2";
import config from "./../../config";

export default class Home extends React.Component {
  static contextType = GAProvider.context;

  static title = "我是首页啊";
  // 数据获取
  static async initData(params: any) {

    return { content: "首页", list: config.tools };
  }

  constructor(props: any) {
    super(props);
  }

  componentDidMount() {}

  render() {
    const { content, list } = this.context;

    return (
      <Layout>
        <div className="container home">
          {/* <BGCanvas2></BGCanvas2> */}
          <div className="tools">
            {list
              ? list.map((item, index) => {
                  return (
                    <div key={"toole"+index}>
                        <p>{item.title}</p>
                        <div className="tool-list">
                            {
                                item.data.map((toolItem,index)=>{
                                    return (
                                      <ToolItem {...toolItem} key={"toolitem" + index}></ToolItem>
                                    )
                                })
                            }
                        </div>
                    </div>
                  );
                })
              : null}
          </div>
        </div>
      </Layout>
    );
  }
}
