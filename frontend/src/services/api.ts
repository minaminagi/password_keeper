import {
    CreateItem,
    DeleteItem,
    GetItem,
    GetVaultMeta,
    InitVault,
    IsVaultInitialized,
    ListItems,
    LockVault,
    RecoverVault,
    UnlockVault,
    UpdateItem,
} from "../../bindings/passwordkeeper/internal/transport/handler";

export type CreateItemPayload = {
    title: string;
    username: string;
    password: string;
    website: string;
    notes: string;
    category: string;
    favorite: boolean;
    tags: string[];
};

export type ListItemsFilter = {
    keyword?: string;
    tag?: string;
    favorite?: boolean | null;
    category?: string;
};

export const api = {
    isVaultInitialized: () => IsVaultInitialized(),
    initVault: (vault_name: string, master_password: string) =>
        InitVault({ vault_name, master_password }),
    unlockVault: (master_password: string) => UnlockVault({ master_password }),
    recoverVault: (recovery_code: string) => RecoverVault({ recovery_code }),
    lockVault: () => LockVault(),
    getVaultMeta: () => GetVaultMeta(),
    listItems: (filter: ListItemsFilter = {}) =>
        ListItems({
            keyword: filter.keyword ?? "",
            tag: filter.tag ?? "",
            favorite: filter.favorite ?? null,
            category: filter.category ?? "",
        }),

    createItem: (payload: CreateItemPayload) => CreateItem(payload),
    updateItem: (payload: CreateItemPayload & { id: string }) =>
        UpdateItem(payload),
    deleteItem: (id: string) => DeleteItem(id),
    getItem: (id: string) => GetItem(id),
};
