import { useState } from "react";
import {
    ActionIcon,
    Alert,
    Button,
    Code,
    Group,
    Stack,
    Text,
} from "@mantine/core";
import {
    IconAlertTriangle,
    IconChevronDown,
    IconChevronUp,
    IconX,
} from "@tabler/icons-react";
import type { AppError } from "../utils/errors";

type Props = {
    error: AppError;
    onClose?: () => void;
};

export function ErrorNotice({ error, onClose }: Props) {
    const [opened, setOpened] = useState(false);

    return (
        <Alert
            className="error-notice"
            color="red"
            icon={<IconAlertTriangle size={20} />}
            title={
                <Group justify="space-between" gap="sm">
                    <Text fw={800}>{error.title}</Text>
                    {onClose && (
                        <ActionIcon
                            aria-label="关闭错误提示"
                            color="red"
                            size="sm"
                            variant="subtle"
                            onClick={onClose}
                        >
                            <IconX size={16} />
                        </ActionIcon>
                    )}
                </Group>
            }
        >
            <Stack gap="xs">
                <Text>{error.message}</Text>
                {error.detail !== error.message && (
                    <>
                        <Button
                            color="red"
                            justify="flex-start"
                            leftSection={
                                opened ? (
                                    <IconChevronUp size={16} />
                                ) : (
                                    <IconChevronDown size={16} />
                                )
                            }
                            p={0}
                            size="compact-sm"
                            variant="transparent"
                            onClick={() => setOpened((value) => !value)}
                        >
                            {opened ? "收起技术详情" : "查看技术详情"}
                        </Button>
                        {opened && (
                            <Code block className="error-detail">
                                {error.detail}
                            </Code>
                        )}
                    </>
                )}
            </Stack>
        </Alert>
    );
}
