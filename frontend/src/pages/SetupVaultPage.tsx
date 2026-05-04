import { useState } from "react";
import {
    Badge,
    Button,
    Group,
    Paper,
    PasswordInput,
    Stack,
    Text,
    TextInput,
    ThemeIcon,
    Title,
} from "@mantine/core";
import {
    IconDatabasePlus,
    IconFingerprint,
    IconLock,
    IconShieldCheck,
} from "@tabler/icons-react";
import { ErrorNotice } from "../components/ErrorNotice";
import { api } from "../services/api";
import { toAppError, type AppError } from "../utils/errors";

type Props = {
    error: AppError | null;
    onCreated: () => void;
};

export function SetupVaultPage({ error, onCreated }: Props) {
    const [vaultName, setVaultName] = useState("Personal Vault");
    const [masterPassword, setMasterPassword] = useState("");
    const [localError, setLocalError] = useState<AppError | null>(error);
    const [submitting, setSubmitting] = useState(false);

    async function handleSubmit(event: React.FormEvent) {
        event.preventDefault();
        setSubmitting(true);
        setLocalError(null);

        try {
            await api.initVault(vaultName, masterPassword);
            onCreated();
        } catch (err) {
            setLocalError(toAppError(err, "创建保险库失败"));
        } finally {
            setSubmitting(false);
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

                            {localError && (
                                <ErrorNotice
                                    error={localError}
                                    onClose={() => setLocalError(null)}
                                />
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
                </Paper>
            </section>
        </main>
    );
}
