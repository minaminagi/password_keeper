import { useEffect, useState } from "react";
import {
    Center,
    Loader,
    Paper,
    Stack,
    Text,
    ThemeIcon,
    Title,
} from "@mantine/core";
import { IconShieldLock } from "@tabler/icons-react";
import { api } from "./services/api";
import { SetupVaultPage } from "./pages/SetupVaultPage";
import { UnlockPage } from "./pages/UnlockPage";
import { VaultPage } from "./pages/VaultPage";
import { toAppError, type AppError } from "./utils/errors";

type AppStage = "loading" | "setup" | "unlock" | "vault";

function App() {
    const [stage, setStage] = useState<AppStage>("loading");
    const [error, setError] = useState<AppError | null>(null);

    useEffect(() => {
        let alive = true;

        api.isVaultInitialized()
            .then((initialized) => {
                if (!alive) return;
                setStage(initialized ? "unlock" : "setup");
            })
            .catch((err) => {
                if (!alive) return;
                setError(toAppError(err, "读取保险库状态失败"));
                setStage("setup");
            });

        return () => {
            alive = false;
        };
    }, []);

    if (stage === "loading") {
        return (
            <main className="auth-screen">
                <Center mih="100vh">
                    <Paper className="loading-card" shadow="xl">
                        <Stack align="center" gap="md">
                            <ThemeIcon size={58} radius={20} variant="gradient">
                                <IconShieldLock size={30} />
                            </ThemeIcon>
                            <Title order={2}>Password Keeper</Title>
                            <Text c="dimmed">正在检查本地保险库状态</Text>
                            <Loader color="teal" type="bars" />
                        </Stack>
                    </Paper>
                </Center>
            </main>
        );
    }

    if (stage === "setup") {
        return (
            <SetupVaultPage error={error} onCreated={() => setStage("vault")} />
        );
    }

    if (stage === "unlock") {
        return (
            <UnlockPage error={error} onUnlocked={() => setStage("vault")} />
        );
    }

    return <VaultPage onLocked={() => setStage("unlock")} />;
}

export default App;
