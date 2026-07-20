const ACTION_LABEL: Record<string, string> = {
  create: "创建",
  save: "保存",
  publish: "发布",
  update: "更新",
  current: "当前编辑",
  restore: "恢复",
};

export function formatVersionTime(iso: string) {
  try {
    return new Date(iso).toLocaleString("zh-CN");
  } catch {
    return iso;
  }
}

export function versionActionLabel(action: string) {
  return ACTION_LABEL[action] || action;
}
