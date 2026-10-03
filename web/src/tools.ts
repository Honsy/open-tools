export const tools: Record<string, { name: string; desc: string; legacy: string }> = {
  prettier: {
    name: "格式化",
    desc: "把 JSON、代码整理成可读的样子。",
    legacy: "legacy/client/pages/prettier",
  },
  crypto: {
    name: "加密解密",
    desc: "摘要、编码和简单加解密。",
    legacy: "legacy 里的 crypto 页面",
  },
  hexconvert: {
    name: "进制转换",
    desc: "二、八、十、十六进制互转。",
    legacy: "legacy 里的 hexconvert 页面",
  },
  moment: {
    name: "时间戳",
    desc: "时间和 Unix 时间戳互转。",
    legacy: "legacy 里的 moment 页面",
  },
  rgb: {
    name: "颜色",
    desc: "颜色值换算。",
    legacy: "legacy 里的 rgb 页面",
  },
  calculator: {
    name: "计算器",
    desc: "四则运算。",
    legacy: "legacy 里的 calculator 页面",
  },
  protobuf: {
    name: "Protobuf",
    desc: "查看 Protobuf 数据。",
    legacy: "legacy 里的 protobuf 页面",
  },
  rmbconvert: {
    name: "人民币大写",
    desc: "金额转中文大写。",
    legacy: "legacy 里的 rmbconvert 页面",
  },
};
