export interface LoginRequest {
    email: string;
    password: string;
}
export interface RegisterRequest {
    name: string;
    email: string;
    password: string;
}
export interface AuthResponse {
    access_token: string;
    refresh_token: string;
    user: {
        id: string;
        email: string;
        name: string;
    };
}
export declare const authApi: {
    login: (data: LoginRequest) => Promise<import("axios").AxiosResponse<AuthResponse, any, {}>>;
    register: (data: RegisterRequest) => Promise<import("axios").AxiosResponse<AuthResponse, any, {}>>;
    logout: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
    me: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const companyApi: {
    list: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const employeeApi: {
    list: (params?: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const groupApi: {
    list: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const floorApi: {
    list: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const departmentApi: {
    list: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const sectionApi: {
    list: (departmentId?: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const designationApi: {
    list: (sectionId?: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const lineApi: {
    list: (sectionId?: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const shiftApi: {
    list: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const attendanceApi: {
    list: (params?: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    clockIn: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    clockOut: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    jobCard: (params?: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    stats: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
    deleteAll: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
    missing: (params: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    absent: (params: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    summary: (params?: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    overtime: (params?: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    overtimeSummary: (params?: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const dataLogApi: {
    import: (data: {
        file_path?: string;
        start_date?: string;
        end_date?: string;
    }) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    list: (params?: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    process: (data: {
        date?: string;
        start_date?: string;
        end_date?: string;
        company_id: string;
    }) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    stats: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
    deleteAll: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const divisionApi: {
    list: () => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const districtApi: {
    list: (divisionId?: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const upazilaApi: {
    list: (districtId?: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const unionApi: {
    list: (upazilaId?: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const requirementApi: {
    list: (params?: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const separationApi: {
    list: (params?: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const idCardApi: {
    list: (params?: Record<string, string>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    get: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    create: (data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    update: (id: string, data: Record<string, unknown>) => Promise<import("axios").AxiosResponse<any, any, {}>>;
    delete: (id: string) => Promise<import("axios").AxiosResponse<any, any, {}>>;
};
export declare const uploadApi: {
    file: (file: File) => Promise<import("axios").AxiosResponse<{
        url: string;
        filename: string;
    }, any, {}>>;
};
