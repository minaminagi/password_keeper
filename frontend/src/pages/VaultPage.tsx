import { useEffect, useState } from "react";
import {
    ActionIcon,
    Badge,
    Button,
    Card,
    Checkbox,
    Code,
    Divider,
    Group,
    Paper,
    PasswordInput,
    ScrollArea,
    SegmentedControl,
    Select,
    SimpleGrid,
    Stack,
    Text,
    TextInput,
    Textarea,
    ThemeIcon,
    Title,
    Tooltip,
} from "@mantine/core";
import {
    IconBrandChrome,
    IconCategory,
    IconDatabase,
    IconEye,
    IconEyeOff,
    IconKey,
    IconLock,
    IconLogout,
    IconNotes,
    IconPencil,
    IconPlus,
    IconRefresh,
    IconSearch,
    IconShieldCheck,
    IconStar,
    IconTrash,
    IconUser,
    IconX,
} from "@tabler/icons-react";
import { ErrorNotice } from "../components/ErrorNotice";
import { api, type CreateItemPayload } from "../services/api";
import { toAppError, type AppError } from "../utils/errors";

type Item = {
    id: string;
    title: string;
    username: string;
    password: string;
    website: string;
    notes: string;
    category: string;
    favorite: boolean;
    tags: string[];
    created_at: string;
    updated_at: string;
};

type Props = {
    onLocked: () => void;
};

type VaultView = "form" | "list" | "detail";

const emptyItem: CreateItemPayload = {
    title: "",
    username: "",
    password: "",
    website: "",
    notes: "",
    category: "login",
    favorite: false,
    tags: [],
};

const categoryOptions = [
    { value: "login", label: "登录账号" },
    { value: "bank", label: "金融" },
    { value: "work", label: "工作" },
    { value: "dev", label: "开发" },
    { value: "other", label: "其他" },
];

const maskedPassword = "••••••";

function parseTags(value: string) {
    return value
        .split(",")
        .map((tag) => tag.trim())
        .filter(Boolean);
}

function formatDate(value: string) {
    if (!value) return "未知时间";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleString();
}

export function VaultPage({ onLocked }: Props) {
    const [items, setItems] = useState<Item[]>([]);
    const [selectedId, setSelectedId] = useState<string>("");
    const [form, setForm] = useState<CreateItemPayload>(emptyItem);
    const [tagsInput, setTagsInput] = useState("");
    const [keyword, setKeyword] = useState("");
    const [category, setCategory] = useState<string | null>("");
    const [favoriteFilter, setFavoriteFilter] = useState("all");
    const [error, setError] = useState<AppError | null>(null);
    const [submitting, setSubmitting] = useState(false);
    const [loading, setLoading] = useState(false);
    const [revealedPasswordId, setRevealedPasswordId] = useState("");
    const [editingId, setEditingId] = useState("");
    const [activeView, setActiveView] = useState<VaultView>("list");
    const [passwordPanelOpen, setPasswordPanelOpen] = useState(false);
    const [currentMasterPassword, setCurrentMasterPassword] = useState("");
    const [newMasterPassword, setNewMasterPassword] = useState("");
    const [confirmMasterPassword, setConfirmMasterPassword] = useState("");
    const [newRecoveryCode, setNewRecoveryCode] = useState("");
    const [changingMasterPassword, setChangingMasterPassword] = useState(false);

    const selectedItem =
        items.find((item) => item.id === selectedId) ?? items[0] ?? null;
    const favoriteCount = items.filter((item) => item.favorite).length;
    const tagCount = new Set(items.flatMap((item) => item.tags)).size;
    const passwordVisible =
        selectedItem?.id !== undefined && selectedItem.id === revealedPasswordId;

    async function refresh() {
        setLoading(true);
        setError(null);
        try {
            const nextItems = await api.listItems({
                keyword,
                category: category ?? "",
                favorite: favoriteFilter === "favorite" ? true : null,
            });
            setItems(nextItems);
            if (
                nextItems.length > 0 &&
                !nextItems.some((item) => item.id === selectedId)
            ) {
                setSelectedId(nextItems[0].id);
            }
            if (nextItems.length === 0) {
                setSelectedId("");
            }
            if (
                revealedPasswordId &&
                !nextItems.some((item) => item.id === revealedPasswordId)
            ) {
                setRevealedPasswordId("");
            }
        } catch (err) {
            setError(toAppError(err, "刷新条目失败"));
        } finally {
            setLoading(false);
        }
    }

    useEffect(() => {
        refresh();
    }, []);

    function resetForm() {
        setForm(emptyItem);
        setTagsInput("");
        setEditingId("");
    }

    function beginEdit(item: Item) {
        setForm({
            title: item.title,
            username: item.username,
            password: item.password,
            website: item.website,
            notes: item.notes,
            category: item.category || "login",
            favorite: item.favorite,
            tags: item.tags,
        });
        setTagsInput(item.tags.join(", "));
        setEditingId(item.id);
        setRevealedPasswordId("");
        setActiveView("form");
    }

    async function handleSave(event: React.FormEvent) {
        event.preventDefault();
        setSubmitting(true);
        setError(null);

        try {
            const payload = {
                ...form,
                tags: parseTags(tagsInput),
            };

            if (editingId) {
                const updated = await api.updateItem({
                    id: editingId,
                    ...payload,
                });
                resetForm();
                await refresh();
                setSelectedId(updated.id);
                setActiveView("detail");
            } else {
                const created = await api.createItem(payload);
                resetForm();
                await refresh();
                setSelectedId(created.id);
                setActiveView("detail");
            }
        } catch (err) {
            setError(toAppError(err, editingId ? "更新条目失败" : "保存条目失败"));
        } finally {
            setSubmitting(false);
        }
    }

    async function handleDelete(id: string) {
        setError(null);
        try {
            await api.deleteItem(id);
            await refresh();
            setActiveView("list");
        } catch (err) {
            setError(toAppError(err, "删除条目失败"));
        }
    }

    async function handleLock() {
        await api.lockVault();
        onLocked();
    }

    async function handleChangeMasterPassword(event: React.FormEvent) {
        event.preventDefault();
        setChangingMasterPassword(true);
        setError(null);

        try {
            if (newMasterPassword !== confirmMasterPassword) {
                setError({
                    title: "修改主密码失败",
                    message: "两次输入的新主密码不一致。",
                    detail: "new master password confirmation mismatch",
                });
                return;
            }

            const meta = await api.changeMasterPassword({
                current_master_password: currentMasterPassword,
                new_master_password: newMasterPassword,
            });
            setNewRecoveryCode(meta.recovery_code);
            setCurrentMasterPassword("");
            setNewMasterPassword("");
            setConfirmMasterPassword("");
        } catch (err) {
            setError(toAppError(err, "修改主密码失败"));
        } finally {
            setChangingMasterPassword(false);
        }
    }

    return (
        <main className="vault-screen">
            <header className="vault-header">
                <Group gap="md">
                    <ThemeIcon size={52} radius={18} variant="gradient">
                        <IconShieldCheck size={28} />
                    </ThemeIcon>
                    <div>
                        <Title order={2}>Password Keeper</Title>
                        <Text c="dimmed" size="sm">
                            本地加密密码管理工作台
                        </Text>
                    </div>
                </Group>

                <Group>
                    <Button
                        variant="light"
                        leftSection={<IconKey size={18} />}
                        onClick={() => {
                            setPasswordPanelOpen((opened) => !opened);
                            setNewRecoveryCode("");
                        }}
                    >
                        {passwordPanelOpen ? "收起修改" : "修改主密码"}
                    </Button>
                    <Button
                        variant="light"
                        leftSection={<IconRefresh size={18} />}
                        loading={loading}
                        onClick={refresh}
                    >
                        刷新
                    </Button>
                    <Button
                        color="dark"
                        leftSection={<IconLogout size={18} />}
                        onClick={handleLock}
                    >
                        锁定
                    </Button>
                </Group>
            </header>

            <SimpleGrid
                className="vault-metrics"
                cols={{ base: 1, sm: 3 }}
                spacing="lg"
                mb="lg"
            >
                <Paper className="metric-card">
                    <Text size="sm" c="dimmed">
                        总条目
                    </Text>
                    <Title order={2}>{items.length}</Title>
                </Paper>
                <Paper className="metric-card accent">
                    <Text size="sm" c="dimmed">
                        收藏
                    </Text>
                    <Title order={2}>{favoriteCount}</Title>
                </Paper>
                <Paper className="metric-card">
                    <Text size="sm" c="dimmed">
                        标签
                    </Text>
                    <Title order={2}>{tagCount}</Title>
                </Paper>
            </SimpleGrid>

            {passwordPanelOpen && (
                <Paper className="password-panel" shadow="md">
                    <Group justify="space-between" align="flex-start" mb="md">
                        <div>
                            <Text c="teal" fw={700} size="sm">
                                安全设置
                            </Text>
                            <Title order={3}>修改主密码</Title>
                            <Text c="dimmed" size="sm" mt={4}>
                                修改后会重新加密所有条目，并生成新的恢复码。
                            </Text>
                        </div>
                        <Tooltip label="关闭">
                            <ActionIcon
                                variant="light"
                                color="gray"
                                onClick={() => {
                                    setPasswordPanelOpen(false);
                                    setNewRecoveryCode("");
                                }}
                            >
                                <IconX size={18} />
                            </ActionIcon>
                        </Tooltip>
                    </Group>

                    {newRecoveryCode ? (
                        <Stack gap="md">
                            <Text c="dimmed" size="sm">
                                主密码已修改。新恢复码只显示这一次，旧恢复码已经失效。
                            </Text>
                            <Code block className="recovery-code">
                                {newRecoveryCode}
                            </Code>
                            <Button
                                onClick={() => {
                                    setPasswordPanelOpen(false);
                                    setNewRecoveryCode("");
                                }}
                            >
                                我已保存新恢复码
                            </Button>
                        </Stack>
                    ) : (
                        <form onSubmit={handleChangeMasterPassword}>
                            <SimpleGrid cols={{ base: 1, md: 3 }} spacing="md">
                                <PasswordInput
                                    label="当前主密码"
                                    leftSection={<IconKey size={18} />}
                                    value={currentMasterPassword}
                                    onChange={(event) =>
                                        setCurrentMasterPassword(
                                            event.currentTarget.value,
                                        )
                                    }
                                    required
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
                                    required
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
                                    required
                                />
                            </SimpleGrid>
                            <Group justify="flex-end" mt="md">
                                <Button
                                    type="submit"
                                    loading={changingMasterPassword}
                                    disabled={
                                        !currentMasterPassword ||
                                        !newMasterPassword ||
                                        !confirmMasterPassword
                                    }
                                >
                                    修改主密码
                                </Button>
                            </Group>
                        </form>
                    )}
                </Paper>
            )}

            {error && (
                <div className="vault-error">
                    <ErrorNotice error={error} onClose={() => setError(null)} />
                </div>
            )}

            <SegmentedControl
                className="vault-view-switcher"
                value={activeView}
                onChange={(value) => setActiveView(value as VaultView)}
                data={[
                    { value: "form", label: editingId ? "编辑" : "新增" },
                    { value: "list", label: "列表" },
                    { value: "detail", label: "详情" },
                ]}
            />

            <section className="vault-grid">
                <Paper
                    className={
                        activeView === "form"
                            ? "composer-card vault-panel active"
                            : "composer-card vault-panel"
                    }
                    shadow="md"
                >
                    <form onSubmit={handleSave}>
                        <Stack gap="md">
                            <Group justify="space-between" align="flex-start">
                                <div>
                                    <Text c="teal" fw={700} size="sm">
                                        {editingId ? "编辑条目" : "新增条目"}
                                    </Text>
                                    <Title order={3}>
                                        {editingId ? "更新这个账号" : "保存一个账号"}
                                    </Title>
                                </div>
                                {editingId && (
                                    <Tooltip label="取消编辑">
                                        <ActionIcon
                                            color="gray"
                                            variant="light"
                                            onClick={resetForm}
                                        >
                                            <IconX size={18} />
                                        </ActionIcon>
                                    </Tooltip>
                                )}
                            </Group>

                            <TextInput
                                label="标题"
                                placeholder="GitHub / Gmail / 服务器"
                                leftSection={<IconDatabase size={18} />}
                                value={form.title}
                                onChange={(event) =>
                                    setForm({
                                        ...form,
                                        title: event.currentTarget.value,
                                    })
                                }
                                required
                            />
                            <TextInput
                                label="用户名"
                                placeholder="name@example.com"
                                leftSection={<IconUser size={18} />}
                                value={form.username}
                                onChange={(event) =>
                                    setForm({
                                        ...form,
                                        username: event.currentTarget.value,
                                    })
                                }
                            />
                            <PasswordInput
                                label="密码"
                                placeholder="输入密码"
                                leftSection={<IconKey size={18} />}
                                value={form.password}
                                onChange={(event) =>
                                    setForm({
                                        ...form,
                                        password: event.currentTarget.value,
                                    })
                                }
                            />
                            <TextInput
                                label="网站"
                                placeholder="https://example.com"
                                leftSection={<IconBrandChrome size={18} />}
                                value={form.website}
                                onChange={(event) =>
                                    setForm({
                                        ...form,
                                        website: event.currentTarget.value,
                                    })
                                }
                            />
                            <Select
                                label="分类"
                                leftSection={<IconCategory size={18} />}
                                data={categoryOptions}
                                value={form.category}
                                onChange={(value) =>
                                    setForm({
                                        ...form,
                                        category: value ?? "other",
                                    })
                                }
                            />
                            <TextInput
                                label="标签"
                                description="多个标签用英文逗号分隔"
                                placeholder="personal, email, important"
                                value={tagsInput}
                                onChange={(event) =>
                                    setTagsInput(event.currentTarget.value)
                                }
                            />
                            <Textarea
                                label="备注"
                                placeholder="恢复码、登录说明或额外提示"
                                leftSection={<IconNotes size={18} />}
                                minRows={4}
                                value={form.notes}
                                onChange={(event) =>
                                    setForm({
                                        ...form,
                                        notes: event.currentTarget.value,
                                    })
                                }
                            />
                            <Checkbox
                                label="加入收藏"
                                checked={form.favorite}
                                onChange={(event) =>
                                    setForm({
                                        ...form,
                                        favorite: event.currentTarget.checked,
                                    })
                                }
                            />
                            <Button
                                type="submit"
                                loading={submitting}
                                disabled={!form.title}
                                leftSection={
                                    editingId ? (
                                        <IconPencil size={18} />
                                    ) : (
                                        <IconPlus size={18} />
                                    )
                                }
                                size="md"
                            >
                                {editingId ? "更新条目" : "保存条目"}
                            </Button>
                        </Stack>
                    </form>
                </Paper>

                <Paper
                    className={
                        activeView === "list"
                            ? "list-card vault-panel active"
                            : "list-card vault-panel"
                    }
                    shadow="md"
                >
                    <Stack gap="md">
                        <Group justify="space-between">
                            <div>
                                <Text c="teal" fw={700} size="sm">
                                    条目列表
                                </Text>
                                <Title order={3}>快速查找</Title>
                            </div>
                            <Badge variant="light">{items.length} items</Badge>
                        </Group>

                        <TextInput
                            placeholder="搜索标题"
                            leftSection={<IconSearch size={18} />}
                            value={keyword}
                            onChange={(event) =>
                                setKeyword(event.currentTarget.value)
                            }
                            onKeyDown={(event) => {
                                if (event.key === "Enter") {
                                    refresh();
                                }
                            }}
                        />
                        <Group grow>
                            <Select
                                placeholder="全部分类"
                                data={[
                                    { value: "", label: "全部分类" },
                                    ...categoryOptions,
                                ]}
                                value={category}
                                onChange={setCategory}
                            />
                            <SegmentedControl
                                value={favoriteFilter}
                                onChange={setFavoriteFilter}
                                data={[
                                    { value: "all", label: "全部" },
                                    { value: "favorite", label: "收藏" },
                                ]}
                            />
                        </Group>
                        <Button
                            variant="light"
                            onClick={refresh}
                            loading={loading}
                            leftSection={<IconSearch size={18} />}
                        >
                            应用筛选
                        </Button>

                        <Divider />

                        <ScrollArea h={520} type="auto">
                            <Stack gap="sm">
                                {items.length === 0 ? (
                                    <Paper className="empty-state">
                                        <ThemeIcon
                                            size={58}
                                            radius={20}
                                            variant="light"
                                        >
                                            <IconLock size={30} />
                                        </ThemeIcon>
                                        <Title order={4}>暂无条目</Title>
                                        <Text c="dimmed" ta="center">
                                            在左侧创建第一个账号，或者调整筛选条件。
                                        </Text>
                                    </Paper>
                                ) : (
                                    items.map((item) => (
                                        <Card
                                            key={item.id}
                                            className={
                                                selectedItem?.id === item.id
                                                    ? "item-card active"
                                                    : "item-card"
                                            }
                                            onClick={() => {
                                                setSelectedId(item.id);
                                                setActiveView("detail");
                                            }}
                                            withBorder
                                        >
                                            <Group justify="space-between">
                                                <div>
                                                    <Group gap={8}>
                                                        <Text fw={800}>
                                                            {item.title}
                                                        </Text>
                                                        {item.favorite && (
                                                            <IconStar
                                                                size={16}
                                                                className="star-icon"
                                                            />
                                                        )}
                                                    </Group>
                                                    <Text c="dimmed" size="sm">
                                                        {item.username ||
                                                            "未填写用户名"}
                                                    </Text>
                                                </div>
                                                <Badge variant="light">
                                                    {item.category || "other"}
                                                </Badge>
                                            </Group>
                                        </Card>
                                    ))
                                )}
                            </Stack>
                        </ScrollArea>
                    </Stack>
                </Paper>

                <Paper
                    className={
                        activeView === "detail"
                            ? "detail-card vault-panel active"
                            : "detail-card vault-panel"
                    }
                    shadow="md"
                >
                    {selectedItem ? (
                        <Stack gap="md">
                            <Group justify="space-between" align="flex-start">
                                <div>
                                    <Badge color="teal" variant="light" mb={8}>
                                        {selectedItem.category || "other"}
                                    </Badge>
                                    <Title
                                        className="copyable-text"
                                        order={2}
                                    >
                                        {selectedItem.title}
                                    </Title>
                                    <Text c="dimmed" size="sm">
                                        更新于 {formatDate(selectedItem.updated_at)}
                                    </Text>
                                </div>
                                <Group gap="xs">
                                    <Tooltip label="编辑条目">
                                        <ActionIcon
                                            color="teal"
                                            variant="light"
                                            size="lg"
                                            onClick={() => beginEdit(selectedItem)}
                                        >
                                            <IconPencil size={18} />
                                        </ActionIcon>
                                    </Tooltip>
                                    <Tooltip label="删除条目">
                                        <ActionIcon
                                            color="red"
                                            variant="light"
                                            size="lg"
                                            onClick={() =>
                                                handleDelete(selectedItem.id)
                                            }
                                        >
                                            <IconTrash size={18} />
                                        </ActionIcon>
                                    </Tooltip>
                                </Group>
                            </Group>

                            <Stack gap="md">
                                <Paper className="secret-field">
                                    <Text size="xs" c="dimmed" tt="uppercase">
                                        用户名
                                    </Text>
                                    <Text className="copyable-text" fw={700}>
                                        {selectedItem.username || "-"}
                                    </Text>
                                </Paper>
                                <Paper className="secret-field">
                                    <Text size="xs" c="dimmed" tt="uppercase">
                                        密码
                                    </Text>
                                    <Group className="password-row" gap="xs">
                                        <Text
                                            className={
                                                passwordVisible
                                                    ? "password-plain copyable-text"
                                                    : "password-mask"
                                            }
                                            fw={700}
                                        >
                                            {selectedItem.password
                                                ? passwordVisible
                                                    ? selectedItem.password
                                                    : maskedPassword
                                                : "-"}
                                        </Text>
                                        {selectedItem.password && (
                                            <Tooltip
                                                label={
                                                    passwordVisible
                                                        ? "隐藏原始密码"
                                                        : "显示原始密码"
                                                }
                                            >
                                                <ActionIcon
                                                    color="teal"
                                                    variant="light"
                                                    onClick={() =>
                                                        setRevealedPasswordId(
                                                            passwordVisible
                                                                ? ""
                                                                : selectedItem.id,
                                                        )
                                                    }
                                                >
                                                    {passwordVisible ? (
                                                        <IconEyeOff size={18} />
                                                    ) : (
                                                        <IconEye size={18} />
                                                    )}
                                                </ActionIcon>
                                            </Tooltip>
                                        )}
                                    </Group>
                                </Paper>
                            </Stack>

                            <Paper className="secret-field">
                                <Text size="xs" c="dimmed" tt="uppercase">
                                    网站
                                </Text>
                                <Text className="copyable-text" fw={700}>
                                    {selectedItem.website || "-"}
                                </Text>
                            </Paper>

                            <Paper className="notes-card">
                                <Text size="xs" c="dimmed" tt="uppercase">
                                    备注
                                </Text>
                                <Text className="copyable-text copyable-block">
                                    {selectedItem.notes || "没有备注内容。"}
                                </Text>
                            </Paper>

                            <Group gap="xs">
                                {selectedItem.tags.length === 0 ? (
                                    <Badge variant="outline">无标签</Badge>
                                ) : (
                                    selectedItem.tags.map((tag) => (
                                        <Badge
                                            className="copyable-text"
                                            key={tag}
                                            variant="light"
                                        >
                                            #{tag}
                                        </Badge>
                                    ))
                                )}
                            </Group>
                        </Stack>
                    ) : (
                        <Paper className="empty-state tall">
                            <ThemeIcon size={64} radius={22} variant="light">
                                <IconDatabase size={34} />
                            </ThemeIcon>
                            <Title order={3}>选择一个条目查看详情</Title>
                            <Text c="dimmed" ta="center">
                                详情区会展示账号、网站、备注和标签。
                            </Text>
                        </Paper>
                    )}
                </Paper>
            </section>
        </main>
    );
}
