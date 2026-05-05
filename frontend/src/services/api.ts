import {
    ChangeMasterPassword,
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

export type ChangeMasterPasswordPayload = {
    current_master_password?: string;
    recovery_code?: string;
    new_master_password: string;
};

export const api = {
    isVaultInitialized: () => IsVaultInitialized(),
    initVault: (vault_name: string, master_password: string) =>
        InitVault({ vault_name, master_password }),
    unlockVault: (master_password: string) => UnlockVault({ master_password }),
    recoverVault: (recovery_code: string) => RecoverVault({ recovery_code }),
    changeMasterPassword: (payload: ChangeMasterPasswordPayload) =>
        ChangeMasterPassword({
            current_master_password: payload.current_master_password ?? "",
            recovery_code: payload.recovery_code ?? "",
            new_master_password: payload.new_master_password,
        }),
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
