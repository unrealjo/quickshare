import { Map } from "immutable";

export const msgs: Map<string, string> = Map({
  "stateMgr.cap.fail": "获取captcha id失败",
  "terms.nothingHere": "这里空空如也",
  "browser.upload.del.fail": "删除上传失败",
  "browser.folder.add.fail": "文件夹名不可为空",
  "browser.del.fail": "至少选择一个文件或文件夹",
  "browser.move.fail": "源与目标相同",
  "browser.share.del.fail": "删除共享失败",
  "browser.share.del": "停止共享",
  "browser.share.add": "共享路径",
  "browser.share.title": "共享列表",
  "browser.share.desc": "所有正在共享的路径",
  "browser.upload.title": "上传列表",
  "browser.upload.desc": "正在上传的文件列表",
  "browser.folder.name": "文件夹名",
  "browser.folder.add": "添加文件夹",
  "browser.upload": "上传",
  "browser.delete": "删除",
  "browser.paste": "粘贴",
  "browser.select": "选择",
  "browser.deselect": "不选",
  "browser.selectAll": "选择所有",
  "browser.stop": "停止",
  "browser.location": "位置",
  "browser.item.title": "列表",
  "browser.used": "已用空间",
  "panes.close": "关闭",
  "login.logout.fail": "登出失败",
  "login.username": "用户名",
  "login.captcha": "验证码",
  "login.pwd": "密码",
  "login.login": "登入",
  "login.logout": "登出",
  "settings.pwd.notSame": "两次密码不同",
  "settings.pwd.empty": "密码不能为空",
  "settings.pwd.notChanged": "新老密码不能相同",
  update: "更新",
  "settings.pwd.old": "当前密码",
  "settings.pwd.new1": "新密码",
  "settings.pwd.new2": "再次输入新密码",
  settings: "设置",
  "settings.chooseLan": "选择语言",
  "settings.pwd.update": "更新密码",
  admin: "管理",
  "update.ok": "更新成功",
  "update.fail": "更新失败",
  "delete.fail": "删除失败",
  "delete.ok": "删除成功",
  delete: "删除",
  spaceLimit: "空间上限",
  uploadLimit: "上传速度限制",
  downloadLimit: "下载速度限制",
  "add.fail": "新增失败",
  "add.ok": "新增成功",
  "role.delete.warning": "注意删除角色后，该角色的用户可能不能登入",
  "user.id": "用户ID",
  "user.add": "新增用户",
  "user.name": "用户名",
  "user.role": "角色",
  "user.password": "密码",
  add: "新增",
  "admin.users": "用户列表",
  "role.add": "新增角色",
  "role.name": "角色名字",
  "admin.roles": "角色列表",
  zhCN: "简体中文",
  enUS: "English(US)",
  "move.fail": "移动失败",
  "share.404.title": "没有找到正在共享的文件夹",
  "share.404.desc": "在列表面板可以共享文件夹",
  "upload.404.title": "没有正在上传的文件",
  "upload.404.desc": "在列表面板可以上传文件",
  detail: "详细",
  refresh: "刷新",
  "refresh-hint": "请稍后刷新查看结果",
  "pane.login": "登入",
  "pane.admin": "管理",
  "pane.settings": "设置",
  "logout.confirm": "确定登出吗？",
  unauthed: "未授权动作",
  "err.tooManyUploads": "不可同时上传超过1000个文件",
  "login.role": "角色",
  "user.profile": "用户信息",
  "user.downLimit": "下载速度限制",
  "user.upLimit": "上传速度限制",
  "user.spaceLimit": "空间限制",
  "cfg.siteName": "站点名字",
  "cfg.siteDesc": "站点描述",
  "cfg.bg": "背景设置",
  "cfg.bg.url": "图片链接",
  "cfg.bg.repeat": "重复",
  "cfg.bg.pos": "位置",
  "cfg.bg.align": "对齐",
  "cfg.bg.bgColor": "背景色",
  reset: "重置",
  "bg.url.alert": "图片链接太短或太长",
  "bg.pos.alert": "Position目前仅支持top, bottom, left, right, center",
  "bg.repeat.alert":
    "Repeat目前仅支持repeat-x, repeat-y, repeat, space, round, no-repeat",
  "bg.align.alert": "Align目前仅支持scroll, fixed, local",
  "prefer.theme": "主题",
  "prefer.theme.url": "主题链接",
  "settings.customLan": "自定义语言包",
  "settings.lanPackURL": "语言包链接",
  "op.fail": "操作失败",
  "op.confirm": "你确定执行此操作吗?",
  "control.panelTabs.filesPanel": "文件",
  "control.panelTabs.uploadingsPanel": "上传",
  "control.panelTabs.sharingsPanel": "共享",
  "control.settingsTabs.managementPane": "管理",
  "control.settingsTabs.preferencePane": "设置",
  "upload.add.fail": "有些文件与上传任务冲突，请检查",
  "err.updater": "updater错误",
  "err.uploadMgr": "upload Manager错误",
  "err.server": "服务器端操作失败",
  "err.script.cors": "跨域脚本错误",
  "err.unknown": "未知错误",
  "item.type": "类型",
  "item.type.folder": "文件夹",
  "item.type.file": "文件",
  "item.name": "名称",
  "item.path": "路径",
  "item.modTime": "修改时间",
  "item.size": "大小",
  "item.progress": "进度",
  "item.downloadURL": "下载链接",
  "error.report.title": "报告错误",
  "op.truncate": "清空",
  "op.submit": "提交",
  "op.cancel": "取消",
  "term.time": "时间",
  "breadcrumb.loc": "位置",
  endpoints: "端点",
  "endpoints.root": "根",
  "endpoints.home": "家",
  "state.stopped": "已停止",
  "state.error": "错误",
  usedSpace: "已用空间",
  resetUsedSpace: "重置已用空间",
  "confirm.resetUsedSpace": "此操作可能会需要一些时间, 确认吗?",
  theme: "主题",
  "theme.light": "光白",
  "theme.dark": "暗黑",
  "siteSettings": "网站设置",
  "siteName": "网站名",
  "siteDesc": "网站描述",
  "allowSetBg": "允许用户自定义背景",
  "autoTheme": "自动切换主题",
  "term.enabled": "启用",
  "term.disabled": "关闭",
  "term.dropAnywhere": "把文件在任意处释放",
  "term.search": "搜索",
  "term.results": "结果",
  "term.noResult": "未找到结果",
  "action.go": "前往",
  "hint.keywords": "请输入关键字，以空格分隔"
});
