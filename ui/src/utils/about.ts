import pkg from '../../package.json'

/**
 * 「关于应用」页面展示的信息。
 *
 * 收款码图片放在 ui/public/support/ 目录下（构建后会原样拷贝到站点根目录）：
 *   ui/public/support/wechat.png   ->  /support/wechat.png
 *   ui/public/support/alipay.png   ->  /support/alipay.png
 * 图片不存在时页面会显示占位框，换成自己的图片后无需改代码。
 * 注意目录名不要用 about：会和前端 /about 路由冲突，直接访问 /about 时被静态目录占用。
 */
export const ABOUT_INFO = {
  /** 站点标题为空时显示的应用名 */
  name: 'Nav',
  description: '专为 NAS 与自托管用户打造的轻量级、本地存储私有云导航站。',
  developer: 'nickkk333',
  /** 开发者 GitHub 主页 / 项目仓库地址 */
  developerUrl: 'https://github.com/nickkk333/nav',
  /** 版本号回退值：优先显示服务端 /api/version 返回的构建版本（发版时由 git tag 注入），
   *  接口不可用（例如旧版服务端）时用 ui/package.json 里的版本 */
  version: pkg.version,
  /** 版本号跳转的 GitHub Releases 页面 */
  releaseUrl: 'https://github.com/nickkk333/nav/releases',
  /** 微信 / 支付宝赞赏码（留空则显示占位框，隐藏对应项） */
  wechatQr: `${import.meta.env.BASE_URL}support/wechat.png`,
  alipayQr: `${import.meta.env.BASE_URL}support/alipay.png`,
}

export const ABOUT_SUPPORT_TIP = '如果您觉得这个项目对您有帮助，欢迎请作者喝杯咖啡'
