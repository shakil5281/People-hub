export interface Department {
    id: string;
    name: string;
    name_bn: string;
}
export interface Section {
    id: string;
    name: string;
    name_bn: string;
    department_id: string;
}
export interface Designation {
    id: string;
    name: string;
    name_bn: string;
    section_id: string;
}
export interface Line {
    id: string;
    name: string;
    name_bn: string;
    section_id: string;
}
