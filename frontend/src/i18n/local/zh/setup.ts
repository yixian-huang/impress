export const setup = {
  title: "初始化 Impress",
  subtitle: "创建管理员账号并配置站点基本信息",
  steps: {
    welcome: "欢迎",
    site: "站点信息",
    admin: "管理员",
    content: "内容模式",
    finish: "完成",
  },
  welcome: {
    heading: "欢迎使用 Impress",
    body: "数据库已通过环境变量连接。请按步骤完成首次配置。",
    database: "数据库类型",
    next: "开始配置",
  },
  site: {
    nameZh: "站点名称（中文）",
    nameEn: "站点名称（英文）",
    defaultLocale: "默认语言",
    localeZh: "中文",
    localeEn: "英文",
  },
  admin: {
    username: "管理员用户名",
    password: "密码",
    confirmPassword: "确认密码",
    hint: "至少 8 位，需包含字母和数字",
  },
  content: {
    heading: "选择初始内容",
    blank: "空白站点",
    blankDesc: "仅基础页面与主题，适合个人博客",
    demo: "演示数据",
    demoDesc: "包含示例文章与咨询站内容，适合体验功能",
  },
  actions: {
    back: "上一步",
    next: "下一步",
    finish: "完成安装",
    finishing: "正在安装…",
  },
  success: {
    heading: "安装完成",
    body: "正在跳转到登录页…",
  },
  errors: {
    passwordMismatch: "两次输入的密码不一致",
    siteNameRequired: "请填写至少一种语言的站点名称",
    setupFailed: "安装失败，请稍后重试",
  },
};
