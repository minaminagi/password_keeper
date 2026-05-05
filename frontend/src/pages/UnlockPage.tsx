import { useState } from "react";
import {
    Badge,
    Button,
    Group,
    Paper,
    PasswordInput,
    Progress,
    SegmentedControl,
    Stack,
    Text,
    TextInput,
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

export function UnlockPage({ error, onUnlocked }: Props) {
    const [masterPassword, setMasterPassword] = useState("");
    const [recoveryCode, setRecoveryCode] = useState("");
    const [unlockMode, setUnlockMode] = useState<"password" | "recovery">(
        "password",
    );
    const [localError, setLocalError] = useState<AppError | null>(error);
    const [submitting, setSubmitting] = useState(false);

    async function handleSubmit(event: React.FormEvent) {
        event.preventDefault();
        setSubmitting(true);
        setLocalError(null);

        try {
            if (unlockMode === "recovery") {
                await api.recoverVault(recoveryCode);
            } else {
                await api.unlockVault(masterPassword);
            }
            onUnlocked();
        } catch (err) {
            setLocalError(
                toAppError(
                    err,
                    unlockMode === "recovery"
                        ? "恢复码解锁失败"
                        : "解锁保险库失败",
                ),
            );
        } finally {
            setSubmitting(false);
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
                                onChange={(value) =>
                                    setUnlockMode(value as "password" | "recovery")
                                }
                                data={[
                                    { value: "password", label: "主密码" },
                                    { value: "recovery", label: "恢复码" },
                                ]}
                            />

                            {unlockMode === "recovery" ? (
                                <TextInput
                                    label="恢复码"
                                    leftSection={<IconLifebuoy size={18} />}
                                    value={recoveryCode}
                                    onChange={(event) =>
                                        setRecoveryCode(event.currentTarget.value)
                                    }
                                    placeholder="输入初始化时保存的恢复码"
                                    size="md"
                                />
                            ) : (
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
                                    unlockMode === "recovery"
                                        ? !recoveryCode
                                        : !masterPassword
                                }
                                size="md"
                                fullWidth
                                rightSection={<IconLockOpen size={18} />}
                            >
                                {unlockMode === "recovery"
                                    ? "用恢复码解锁"
                                    : "解锁"}
                            </Button>
                        </Stack>
                    </form>
                </Paper>
            </section>
        </main>
    );
}
