import { useState } from "react";
import {
    Badge,
    Button,
    Code,
    Group,
    Paper,
    PasswordInput,
    Stack,
    Text,
    TextInput,
    Textarea,
    ThemeIcon,
    Title,
} from "@mantine/core";
import {
    IconDatabasePlus,
    IconFingerprint,
    IconLifebuoy,
    IconLock,
    IconShieldCheck,
    IconUpload,
} from "@tabler/icons-react";
import { ErrorNotice } from "../components/ErrorNotice";
import { api } from "../services/api";
import { toAppError, type AppError } from "../utils/errors";

type Props = {
    error: AppError | null;
    onCreated: () => void;
    onImported: () => void;
};

export function SetupVaultPage({ error, onCreated, onImported }: Props) {
    const [vaultName, setVaultName] = useState("Personal Vault");
    const [masterPassword, setMasterPassword] = useState("");
    const [backupCipherText, setBackupCipherText] = useState("");
    const [backupPassword, setBackupPassword] = useState("");
    const [localError, setLocalError] = useState<AppError | null>(error);
    const [submitting, setSubmitting] = useState(false);
    const [importing, setImporting] = useState(false);
    const [importPanelOpen, setImportPanelOpen] = useState(false);
    const [recoveryCode, setRecoveryCode] = useState("");

    async function handleSubmit(event: React.FormEvent) {
        event.preventDefault();
        setSubmitting(true);
        setLocalError(null);

        try {
            const meta = await api.initVault(vaultName, masterPassword);
            setRecoveryCode(meta.recovery_code);
        } catch (err) {
            setLocalError(toAppError(err, "创建保险库失败"));
        } finally {
            setSubmitting(false);
        }
    }

    async function handleImportBackup() {
        setImporting(true);
        setLocalError(null);

        try {
            await api.importBackup(backupCipherText, backupPassword);
            setBackupCipherText("");
            setBackupPassword("");
            onImported();
        } catch (err) {
            setLocalError(toAppError(err, "导入密文备份失败"));
        } finally {
            setImporting(false);
        }
    }

    return (
        <main className="auth-screen">
            <section className="auth-grid">
                <div className="auth-hero">
                    <Badge variant="light" color="teal" size="lg">
                        本地优先 · 加密存储
                    </Badge>
                    <Title className="auth-title">
                        创建你的离线密码保险库
                    </Title>
                    <Text className="auth-copy">
                        主密码不会离开本机。项目后端负责密钥派生、加密和
                        SQLite 持久化，前端只负责清晰、安全地组织操作入口。
                    </Text>

                    <div className="security-orbit">
                        <ThemeIcon className="orbit-core" size={96} radius={32}>
                            <IconShieldCheck size={48} />
                        </ThemeIcon>
                        <span className="orbit-dot dot-a">
                            <IconLock size={18} />
                        </span>
                        <span className="orbit-dot dot-b">
                            <IconFingerprint size={18} />
                        </span>
                        <span className="orbit-dot dot-c">
                            <IconDatabasePlus size={18} />
                        </span>
                    </div>
                </div>

                <Paper className="auth-card" shadow="xl">
                    {recoveryCode ? (
                        <Stack gap="lg">
                            <ThemeIcon size={62} radius={22} variant="gradient">
                                <IconLifebuoy size={34} />
                            </ThemeIcon>
                            <div>
                                <Text c="teal" fw={700} size="sm">
                                    只显示这一次
                                </Text>
                                <Title order={2}>保存恢复码</Title>
                                <Text c="dimmed" mt={6}>
                                    如果忘记主密码，只能用这个恢复码解锁保险库。
                                    关闭此页面后，应用不会再次显示它。
                                </Text>
                            </div>

                            <Code block className="recovery-code">
                                {recoveryCode}
                            </Code>

                            <Text c="red" size="sm" fw={700}>
                                请把恢复码保存在安全的位置。丢失主密码和恢复码后，
                                已加密的数据无法恢复。
                            </Text>

                            <Button
                                size="md"
                                rightSection={<IconShieldCheck size={18} />}
                                onClick={onCreated}
                            >
                                我已保存恢复码，进入保险库
                            </Button>
                        </Stack>
                    ) : (
                        <Stack gap="lg">
                            <form onSubmit={handleSubmit}>
                                <Stack gap="lg">
                                    <div>
                                        <Text c="teal" fw={700} size="sm">
                                            第一次使用
                                        </Text>
                                        <Title order={2}>初始化保险库</Title>
                                        <Text c="dimmed" mt={6}>
                                            设置名称和主密码后即可开始保存账号条目。
                                        </Text>
                                    </div>

                                    <TextInput
                                        label="保险库名称"
                                        leftSection={<IconDatabasePlus size={18} />}
                                        value={vaultName}
                                        onChange={(event) =>
                                            setVaultName(event.currentTarget.value)
                                        }
                                        placeholder="Personal Vault"
                                        size="md"
                                    />
                                    <PasswordInput
                                        label="主密码"
                                        leftSection={<IconLock size={18} />}
                                        value={masterPassword}
                                        onChange={(event) =>
                                            setMasterPassword(event.currentTarget.value)
                                        }
                                        placeholder="输入一个足够强的主密码"
                                        size="md"
                                    />

                                    <Button
                                        type="button"
                                        variant="light"
                                        leftSection={<IconUpload size={18} />}
                                        onClick={() => {
                                            setImportPanelOpen((opened) => !opened);
                                            setLocalError(null);
                                        }}
                                    >
                                        {importPanelOpen
                                            ? "收起导入"
                                            : "导入已有密文备份"}
                                    </Button>

                                    {importPanelOpen && (
                                        <Stack gap="md">
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
                                                leftSection={<IconLock size={18} />}
                                                value={backupPassword}
                                                onChange={(event) =>
                                                    setBackupPassword(
                                                        event.currentTarget.value,
                                                    )
                                                }
                                                placeholder="输入导出备份时设置的密码"
                                            />
                                            <Button
                                                type="button"
                                                variant="filled"
                                                loading={importing}
                                                disabled={
                                                    !backupCipherText ||
                                                    !backupPassword
                                                }
                                                leftSection={<IconUpload size={18} />}
                                                onClick={handleImportBackup}
                                            >
                                                导入密文备份
                                            </Button>
                                        </Stack>
                                    )}

                                    <Group justify="space-between" align="center">
                                        <Text size="sm" c="dimmed">
                                            建议不少于 12 位，并混合字符类型。
                                        </Text>
                                        <Button
                                            type="submit"
                                            loading={submitting}
                                            disabled={!vaultName || !masterPassword}
                                            size="md"
                                            rightSection={<IconShieldCheck size={18} />}
                                        >
                                            创建保险库
                                        </Button>
                                    </Group>
                                </Stack>
                            </form>

                            {localError && (
                                <ErrorNotice
                                    error={localError}
                                    onClose={() => setLocalError(null)}
                                />
                            )}

                        </Stack>
                    )}
                </Paper>
            </section>
        </main>
    );
}
