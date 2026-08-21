export interface JobCardRecord {
    id: number;
    employee: string;
    employeeId: string;
    company: string;
    department: string;
    designation: string;
    line: string;
    section: string;
    group: string;
    date: string;
    shift: string;
    inTime: string;
    outTime: string;
    late: string;
    overTime: string;
    status: "Present" | "Late" | "Absent" | "Half Day" | "Holiday" | "Leave";
}
export interface EmployeeInfo {
    employee: string;
    employeeId: string;
    company: string;
    department: string;
    designation: string;
    line: string;
    section: string;
    group: string;
}
export declare function getJobCardRecords(): JobCardRecord[];
export declare function getEmployees(): EmployeeInfo[];
export declare const companyOptions: {
    value: string;
    label: string;
}[];
export declare const departmentOptions: {
    value: string;
    label: string;
}[];
export declare const designationOptions: {
    value: string;
    label: string;
}[];
export declare const lineOptions: {
    value: string;
    label: string;
}[];
export declare const sectionOptions: {
    value: string;
    label: string;
}[];
export declare const groupOptions: {
    value: string;
    label: string;
}[];
