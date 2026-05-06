import { useState } from "react";
import {
    Badge,
    Button,
    Code,
    Group,
    Paper,
    PasswordInput,
    Progress,
    SegmentedControl,
    Stack,
    Text,
    TextInput,
    Textarea,
    ThemeIcon,
    Title,
} from "@mantine/core";
import {
    IconKey,
    IconLifebuoy,
    IconLockOpen,
    IconShieldLock,
    IconSparkles,
} from "@tabler/icons-react";
import { ErrorNotice } from "../components/ErrorNotice";
import { api } from "../services/api";
import { toAppError, type AppError } from "../utils/errors";

type Props = {
    error: AppError | null;
    onUnlocked: () => void;
};

type UnlockMode = "password" | "reset" | "import";

export function UnlockPage({ error, onUnlocked }: Props) {
    const [masterPassword, setMasterPassword] = useState("");
    const [recoveryCode, setRecoveryCode] = useState("");
    const [newMasterPassword, setNewMasterPassword] = useState("");
    const [confirmMasterPassword, setConfirmMasterPassword] = useState("");
    const [backupCipherText, setBackupCipherText] = useState("");
    const [backupPassword, setBackupPassword] = useState("");
    const [newRecoveryCode, setNewRecoveryCode] = useState("");
    const [backupImported, setBackupImported] = useState(false);
    const [unlockMode, setUnlockMode] = useState<UnlockMode>("password");
    const [localError, setLocalError] = useState<AppError | null>(error);
    const [submitting, setSubmitting] = useState(false);

    async function handleSubmit(event: React.FormEvent) {
        event.preventDefault();
        setSubmitting(true);
        setLocalError(null);

        try {
            if (unlockMode === "import") {
                await handleImportBackup();
                return;
            }
            if (unlockMode === "reset") {
                if (newMasterPassword !== confirmMasterPassword) {
                    setLocalError({
                        title: "重置主密码失败",
                        message: "两次输入的新主密码不一致。",
                        detail: "new master password confirmation mismatch",
                    });
                    return;
                }
                const meta = await api.changeMasterPassword({
                    recovery_code: recoveryCode,
                    new_master_password: newMasterPassword,
                });
                setNewRecoveryCode(meta.recovery_code);
                return;
            }
            await api.unlockVault(masterPassword);
            onUnlocked();
        } catch (err) {
            setLocalError(
                toAppError(
                        err,
                    unlockMode === "reset"
                        ? "重置主密码失败"
                        : unlockMode === "import"
                        ? "导入密文备份失败"
                        : "解锁保险库失败",
                ),
            );
        } finally {
            setSubmitting(false);
        }
    }

    async function handleImportBackup() {
        setBackupImported(false);
        setLocalError(null);

        try {
            await api.importBackup(backupCipherText, backupPassword);
            setBackupCipherText("");
            setBackupPassword("");
            setMasterPassword("");
            setBackupImported(true);
            setUnlockMode("password");
        } catch (err) {
            setLocalError(toAppError(err, "导入密文备份失败"));
        }
    }

    return (
        <main className="auth-screen">
            <section className="auth-grid compact">
                <div className="auth-hero">
                    <Badge variant="light" color="cyan" size="lg">
                        Zero knowledge workspace
                    </Badge>
                    <Title className="auth-title">解锁你的密码工作台</Title>
                    <Text className="auth-copy">
                        输入主密码后，后端会重新派生密钥并解密 vault
                        校验块。校验成功后，前端才会进入条目管理界面。
                    </Text>
                    <Paper className="signal-panel">
                        <Group justify="space-between" mb="sm">
                            <Text size="sm" fw={700}>
                                本地安全状态
                            </Text>
                            <IconSparkles size={18} />
                        </Group>
                        <Progress value={82} color="teal" radius="xl" />
                        <Text size="xs" c="dimmed" mt="sm">
                            加密数据仅保存在本地 SQLite 数据库。
                        </Text>
                    </Paper>
                </div>

                <Paper className="auth-card" shadow="xl">
                    <form onSubmit={handleSubmit}>
                        <Stack gap="lg">
                            <ThemeIcon size={62} radius={22} variant="gradient">
                                <IconShieldLock size={34} />
                            </ThemeIcon>
                            <div>
                                <Text c="teal" fw={700} size="sm">
                                    欢迎回来
                                </Text>
                                <Title order={2}>解锁保险库</Title>
                            </div>

                            <SegmentedControl
                                value={unlockMode}
                                onChange={(value) => {
                                    setUnlockMode(value as UnlockMode);
                                    setLocalError(null);
                                    setNewRecoveryCode("");
                                    setBackupImported(false);
                                }}
                                data={[
                                    { value: "password", label: "主密码" },
                                    { value: "reset", label: "重置密码" },
                                    { value: "import", label: "导入备份" },
                                ]}
                            />

                            {newRecoveryCode ? (
                                <Stack gap="sm">
                                    <Text fw={700}>主密码已更新</Text>
                                    <Text c="dimmed" size="sm">
                                        新恢复码只显示这一次。旧恢复码已经失效，请保存下面的新恢复码。
                                    </Text>
                                    <Code block className="recovery-code">
                                        {newRecoveryCode}
                                    </Code>
                                </Stack>
                            ) : unlockMode === "reset" ? (
                                <Stack gap="md">
                                    <TextInput
                                        label="恢复码"
                                        leftSection={<IconLifebuoy size={18} />}
                                        value={recoveryCode}
                                        onChange={(event) =>
                                            setRecoveryCode(
                                                event.currentTarget.value,
                                            )
                                        }
                                        placeholder="输入初始化时保存的恢复码"
                                        size="md"
                                    />
                                    <PasswordInput
                                        label="新主密码"
                                        leftSection={<IconKey size={18} />}
                                        value={newMasterPassword}
                                        onChange={(event) =>
                                            setNewMasterPassword(
                                                event.currentTarget.value,
                                            )
                                        }
                                        placeholder="输入新的主密码"
                                        size="md"
                                    />
                                    <PasswordInput
                                        label="确认新主密码"
                                        leftSection={<IconKey size={18} />}
                                        value={confirmMasterPassword}
                                        onChange={(event) =>
                                            setConfirmMasterPassword(
                                                event.currentTarget.value,
                                            )
                                        }
                                        placeholder="再次输入新的主密码"
                                        size="md"
                                    />
                                </Stack>
                            ) : unlockMode === "import" ? (
                                <Stack gap="md">
                                    <Text c="dimmed" size="sm">
                                        导入会覆盖当前本地数据库。导入成功后，请使用导入库的主密码解锁。
                                    </Text>
                                    <Textarea
                                        label="密文备份"
                                        description="粘贴以 PKB1. 开头的导出文本"
                                        minRows={5}
                                        autosize
                                        value={backupCipherText}
                                        onChange={(event) =>
                                            setBackupCipherText(
                                                event.currentTarget.value,
                                            )
                                        }
                                        placeholder="PKB1..."
                                    />
                                    <PasswordInput
                                        label="导出密码"
                                        leftSection={<IconKey size={18} />}
                                        value={backupPassword}
                                        onChange={(event) =>
                                            setBackupPassword(
                                                event.currentTarget.value,
                                            )
                                        }
                                        placeholder="输入导出备份时设置的密码"
                                    />
                                </Stack>
                            ) : (
                                <Stack gap="md">
                                    {backupImported && (
                                        <Text c="teal" size="sm" fw={700}>
                                            备份已导入。请使用导入库的主密码解锁。
                                        </Text>
                                    )}
                                    <PasswordInput
                                        label="主密码"
                                        leftSection={<IconKey size={18} />}
                                        value={masterPassword}
                                        onChange={(event) =>
                                            setMasterPassword(
                                                event.currentTarget.value,
                                            )
                                        }
                                        placeholder="输入主密码"
                                        size="md"
                                    />
                                </Stack>
                            )}

                            {localError && (
                                <ErrorNotice
                                    error={localError}
                                    onClose={() => setLocalError(null)}
                                />
                            )}

                            <Button
                                type="submit"
                                loading={submitting}
                                disabled={
                                    newRecoveryCode
                                        ? false
                                        : unlockMode === "reset"
                                          ? !recoveryCode ||
                                            !newMasterPassword ||
                                            !confirmMasterPassword
                                          : unlockMode === "import"
                                          ? !backupCipherText || !backupPassword
                                          : !masterPassword
                                }
                                size="md"
                                fullWidth
                                rightSection={<IconLockOpen size={18} />}
                                onClick={
                                    newRecoveryCode
                                        ? (event) => {
                                              event.preventDefault();
                                              onUnlocked();
                                          }
                                        : undefined
                                }
                            >
                                {newRecoveryCode
                                    ? "我已保存新恢复码，进入保险库"
                                    : unlockMode === "reset"
                                    ? "用恢复码重置主密码"
                                    : unlockMode === "import"
                                    ? "覆盖导入备份"
                                    : "解锁"}
                            </Button>
                        </Stack>
                    </form>
                </Paper>
            </section>
        </main>
    );
}
