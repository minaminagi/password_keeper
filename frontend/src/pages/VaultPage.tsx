import { useEffect, useState } from "react";
import {
    ActionIcon,
    Badge,
    Button,
    Card,
    Checkbox,
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
    IconKey,
    IconLock,
    IconLogout,
    IconNotes,
    IconPlus,
    IconRefresh,
    IconSearch,
    IconShieldCheck,
    IconStar,
    IconTrash,
    IconUser,
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

    const selectedItem =
        items.find((item) => item.id === selectedId) ?? items[0] ?? null;
    const favoriteCount = items.filter((item) => item.favorite).length;
    const tagCount = new Set(items.flatMap((item) => item.tags)).size;

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
        } catch (err) {
            setError(toAppError(err, "刷新条目失败"));
        } finally {
            setLoading(false);
        }
    }

    useEffect(() => {
        refresh();
    }, []);

    async function handleCreate(event: React.FormEvent) {
        event.preventDefault();
        setSubmitting(true);
        setError(null);

        try {
            const created = await api.createItem({
                ...form,
                tags: parseTags(tagsInput),
            });
            setForm(emptyItem);
            setTagsInput("");
            await refresh();
            setSelectedId(created.id);
        } catch (err) {
            setError(toAppError(err, "保存条目失败"));
        } finally {
            setSubmitting(false);
        }
    }

    async function handleDelete(id: string) {
        setError(null);
        try {
            await api.deleteItem(id);
            await refresh();
        } catch (err) {
            setError(toAppError(err, "删除条目失败"));
        }
    }

    async function handleLock() {
        await api.lockVault();
        onLocked();
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

            <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="lg" mb="lg">
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

            {error && (
                <div className="vault-error">
                    <ErrorNotice error={error} onClose={() => setError(null)} />
                </div>
            )}

            <section className="vault-grid">
                <Paper className="composer-card" shadow="md">
                    <form onSubmit={handleCreate}>
                        <Stack gap="md">
                            <div>
                                <Text c="teal" fw={700} size="sm">
                                    新增条目
                                </Text>
                                <Title order={3}>保存一个账号</Title>
                            </div>

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
                                leftSection={<IconPlus size={18} />}
                                size="md"
                            >
                                保存条目
                            </Button>
                        </Stack>
                    </form>
                </Paper>

                <Paper className="list-card" shadow="md">
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
                                            onClick={() =>
                                                setSelectedId(item.id)
                                            }
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

                <Paper className="detail-card" shadow="md">
                    {selectedItem ? (
                        <Stack gap="md">
                            <Group justify="space-between" align="flex-start">
                                <div>
                                    <Badge color="teal" variant="light" mb={8}>
                                        {selectedItem.category || "other"}
                                    </Badge>
                                    <Title order={2}>{selectedItem.title}</Title>
                                    <Text c="dimmed" size="sm">
                                        更新于 {formatDate(selectedItem.updated_at)}
                                    </Text>
                                </div>
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

                            <SimpleGrid cols={2}>
                                <Paper className="secret-field">
                                    <Text size="xs" c="dimmed" tt="uppercase">
                                        用户名
                                    </Text>
                                    <Text fw={700}>
                                        {selectedItem.username || "-"}
                                    </Text>
                                </Paper>
                                <Paper className="secret-field">
                                    <Text size="xs" c="dimmed" tt="uppercase">
                                        密码
                                    </Text>
                                    <Text fw={700}>
                                        {selectedItem.password
                                            ? "••••••••••••"
                                            : "-"}
                                    </Text>
                                </Paper>
                            </SimpleGrid>

                            <Paper className="secret-field">
                                <Text size="xs" c="dimmed" tt="uppercase">
                                    网站
                                </Text>
                                <Text fw={700}>
                                    {selectedItem.website || "-"}
                                </Text>
                            </Paper>

                            <Paper className="notes-card">
                                <Text size="xs" c="dimmed" tt="uppercase">
                                    备注
                                </Text>
                                <Text>
                                    {selectedItem.notes || "没有备注内容。"}
                                </Text>
                            </Paper>

                            <Group gap="xs">
                                {selectedItem.tags.length === 0 ? (
                                    <Badge variant="outline">无标签</Badge>
                                ) : (
                                    selectedItem.tags.map((tag) => (
                                        <Badge key={tag} variant="light">
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
