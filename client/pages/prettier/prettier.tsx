import * as React from "react";
import { Layout } from "../../layout";
import GAProvider from "../../ga-provider";
import prettier from "prettier/standalone";
import parserGraphql from "prettier/parser-graphql";
import babel from "prettier/parser-babel";
// import typescript from "prettier/parser-typescript";
// import flow from "prettier/parser-flow";
import postcss from "prettier/parser-postcss";
import markdown from "prettier/parser-markdown";
import html from "prettier/parser-html";
import yaml from "prettier/parser-yaml";

interface FormatItem {
  title: string;
  type: string;
}

interface PrettierState {
  formatText: string;
  formatedText: string;
  selectIndex: number;
}

export default class Prettier extends React.Component<{}, PrettierState> {
  static contextType = GAProvider.context;

  static title = "格式化工具-OB工具-Prettier"

  static initData() {
    let list: FormatItem[] = [
      { title: "babel", type: "babel"},
      { title: "babel-flow", type: "babel-flow" },
      // { title: "babel-ts", type: "babel-ts" },
      // { title: "flow", type: "flow" },
      // { title: "typescript", type: "typescript" },
      { title: "css", type: "css" },
      { title: "scss", type: "scss" },
      { title: "less", type: "less" },
      { title: "json", type: "json" },
      { title: "json5", type: "json5" },
      { title: "json-stringify", type: "json-stringify" },
      { title: "graphql", type: "graphql" },
      { title: "markdown", type: "markdown" },
      { title: "mdx", type: "mdx" },
      { title: "html", type: "html" },
      { title: "vue", type: "vue" },
      { title: "angular", type: "angular" },
      { title: "lwc", type: "lwc" },
      { title: "yaml", type: "yaml" },
      { title: "特殊化json", type: "sp" },
    ];

    return { list: list };
  }

  formatTextArea: React.RefObject<HTMLTextAreaElement>;
  formatedTextArea: React.RefObject<HTMLTextAreaElement>;

  constructor(props: any) {
    super(props);
    this.formatTextArea = React.createRef();
    this.formatedTextArea = React.createRef();
    this.state = {
      formatText: "",
      formatedText: "",
      selectIndex: 0,
    };
  }
  componentDidMount() {}

//   格式化插件对应配置
  formatConfig = (item:FormatItem) =>{
    if (item.type === "babel" || item.type === "babel-flow" || item.type === "json" || item.type === "json5" || item.type === "json-stringify") {
        return {
            parser: item.type,
            plugins: [babel],
        }
    }else if(item.type==="babel-ts" || item.type === "typescript"){
        return {
            // parser: item.type,
            // plugins: [typescript],
        }
    }else if(item.type==="flow"){
        return {
            // parser: item.type,
            // plugins: [flow],
        }
    }else if(item.type==="css" || item.type==="scss" || item.type==="less"){
        return {
            parser: item.type,
            plugins: [postcss],
        }
    }else if(item.type==="graphql"){
        return {
            parser: item.type,
            plugins: [parserGraphql],
        }
    }else if(item.type==="html" || item.type==="vue" ||item.type==="angular"){
        return {
            parser: item.type,
            plugins: [html],
        }
    }else if(item.type==="graphql"){
        return {
            parser: item.type,
            plugins: [parserGraphql],
        }
    }else if(item.type==="markdown"){
        return {
            parser: item.type,
            plugins: [markdown],
        }
    }else if(item.type==="yaml"){
        return {
            parser: item.type,
            plugins: [yaml],
        }
    }
  }

//   开始格式化
  startPretter = () => {
    const { list } = this.context;
    const { formatText, selectIndex } = this.state;
    let selectItem = list[selectIndex];
    let result
    if (selectItem.type === "sp") {
      result = formatText.replace(/, /g,"\",\"")
              .replace(/=/g,"\":\"")
              .replace(/\:\{/g,"\":{\"")
              .replace(/入参csfParamMap/,"{\"csfParamMap")
              .replace(/入参object/,",\"object")
              .replace(/} ,/g,"\"} ,")
              .replace(/\"\]/g,"]")
              .replace(/\"\[/g,"[")
              .replace(/入参beans\:\[\{/,",\"},\"beans\":\[{\"")
              .replace(/入参beans\:\[/,",\"},\"beans\":[")
              .replace("返回编码：0 返回信息returnMessage：调用EOMS接口成功","}")
              .replace(/]"/g,"]")
              .replace(/\"\[/g,"[")
              .replace(/\"\}\]\ }/g,"\"\"\}\]\}")
              .replace(/\"\}\"/g,"\"\"")
              
    }else{
      result = prettier.format(formatText, this.formatConfig(selectItem));

    }

    console.log(result);
    this.setState({
        formatedText:result
      },()=>{
        this.formatedTextArea.current.style.height = this.formatedTextArea.current.scrollHeight + "px"
      })
  };
  //改变格式化类型
  changeFormatType = (e: any, index: number) => {
      this.setState({
          selectIndex:index
      })
  };

  // 数据接收
  inputValueChange = (key: keyof PrettierState, value: string | number) => {
    this.setState({
      [key]: value,
    } as Pick<PrettierState, keyof PrettierState>);
  };
  render() {
    const { formatText, formatedText, selectIndex } = this.state;
    const { list } = this.context;
    return (
      <Layout>
        <div className="container prettier">
          <div className="formats">
            <h5>请选择以下任意一种</h5>
            {list.map((item, index) => {
              return (
                <span
                  onClick={(e) => this.changeFormatType(e, index)}
                  key={item.title}
                  className={
                    selectIndex !== index ? "tag is-light" : "tag is-dark"
                  }
                >
                  {item.title}
                </span>
              );
            })}
          </div>
          <div className="content">
            <div className="columns">
              <div className="column is-two-fifths">
                <label className="active">待格式化文本</label>
                <textarea
                  ref={this.formatTextArea}
                  value={formatText}
                  style={{ minHeight: 200 }}
                  className="textarea"
                  onChange={(e) =>
                    this.inputValueChange("formatText", e.target.value)
                  }
                ></textarea>
              </div>
              <div className="column center">
                <button className="button" onClick={this.startPretter}>
                  格式化
                </button>
              </div>
              <div className="column is-two-fifths">
                <label className="active">格式化文本</label>
                <textarea
                  ref={this.formatedTextArea}
                  value={formatedText}
                  style={{ minHeight: 200 }}
                  className="textarea"
                  onChange={(e) =>
                    this.inputValueChange("formatedText", e.target.value)
                  }
                ></textarea>
              </div>
            </div>
          </div>
        </div>
      </Layout>
    );
  }
}
