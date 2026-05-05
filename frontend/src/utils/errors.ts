export type AppError = {
    title: string;
    message: string;
    detail: string;
};

const knownMessages: Array<[RegExp, string]> = [
    [/invalid master password/i, "主密码不正确，请重新输入。"],
    [/invalid recovery code/i, "恢复码不正确，请检查后重新输入。"],
    [/vault recovery not found/i, "当前保险库没有可用恢复码。"],
    [/vault is locked/i, "保险库已锁定，请先解锁后再操作。"],
    [/vault meta already exists/i, "保险库已经初始化，不需要重复创建。"],
    [/vault meta not found/i, "还没有创建保险库，请先完成初始化。"],
    [/item title is required/i, "请填写条目标题。"],
    [/items not found/i, "没有找到这个条目，可能已经被删除。"],
    [/tag already exists/i, "标签已存在。"],
    [/FOREIGN KEY constraint failed/i, "标签关联失败，请刷新后重试。"],
    [/UNIQUE constraint failed: tags\.name/i, "标签名称重复，请换一个名称。"],
    [/decrypt:/i, "解密失败，请确认主密码是否正确。"],
    [/unable to connect to frontend server/i, "前端开发服务未启动，请重新运行 wails3 dev。"],
    [/Failed to fetch|NetworkError/i, "无法连接到应用后端，请确认程序仍在运行。"],
];

export function toAppError(err: unknown, fallbackTitle = "操作失败"): AppError {
    const detail = normalizeErrorDetail(err);
    const message =
        knownMessages.find(([pattern]) => pattern.test(detail))?.[1] ??
        simplifyErrorMessage(detail);

    return {
        title: fallbackTitle,
        message,
        detail,
    };
}

function normalizeErrorDetail(err: unknown): string {
    if (err instanceof Error) {
        return err.message || err.name;
    }
    if (typeof err === "string") {
        return err;
    }
    if (err && typeof err === "object") {
        if ("message" in err && typeof err.message === "string") {
            return err.message;
        }
        try {
            return JSON.stringify(err);
        } catch {
            return String(err);
        }
    }
    return String(err);
}

function simplifyErrorMessage(detail: string) {
    const cleaned = detail
        .replace(/^error:\s*/i, "")
        .replace(/^panic error:\s*/i, "")
        .trim();

    const parts = cleaned
        .split(":")
        .map((part) => part.trim())
        .filter(Boolean);

    if (parts.length > 1) {
        return parts[parts.length - 1];
    }
    return cleaned || "发生未知错误，请稍后重试。";
}
