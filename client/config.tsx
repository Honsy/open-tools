import { ToolObj } from "./interface";

const list:ToolObj[] = [
    {
      title: "常用工具",
      data: [
        { title: "计算器", href: "/calculator", color: "#FFB6C1" },
        { title: "人民币大小写转换", href: "/rmbconvert", color: "#8FBC8F" },
      ],
    },
    {
      title: "代码工具",
      data: [
        { title: "加密解密", href: "/crypto", color: "#E6E6FA" },
        { title: "格式化", href: "/prettier", color: "#4169E1" },
        { title: "进制转换", href: "/hexconvert", color: "#5F9EA0" },
        { title: "时间戳转换", href: "/moment", color: "#8FBC8F" },
        { title: "RGB", href: "/rgb", color: "#C0C0C0" },
      ],
    },
    {
      title: "写着玩玩",
      data: [
        { title: "加密解密", href: "/crypto", color: "#E6E6FA" },
      ],
    }
]


export default {
    tools:list
}