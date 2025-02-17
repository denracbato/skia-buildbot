// DO NOT EDIT. This file is automatically generated.

export interface SetNoteRequest {
	Message: string;
}

export interface SupplyChromeOSRequest {
	SSHUserIP: string;
	SuppliedDimensions: SwarmingDimensions;
}

export interface SetAttachedDevice {
	AttachedDevice: AttachedDevice;
}

export interface Annotation {
	Message: string;
	User: string;
	Timestamp: string;
}

export interface CacheRequest {
	Name: string;
	Path: string;
}

export interface Package {
	name: string;
	path: string;
	version: string;
}

export interface TaskRequest {
	Caches: CacheRequest[];
	CasInput: string;
	CipdPackages: Package[];
	Command: string[];
	Dimensions: string[];
	Env: { [key: string]: string };
	EnvPrefixes: { [key: string]: string[] };
	ExecutionTimeout: Duration;
	Expiration: Duration;
	ExtraArgs: string[];
	Idempotent: boolean;
	IoTimeout: Duration;
	Name: string;
	Outputs: string[];
	ServiceAccount: string;
	Tags: string[];
	TaskSchedulerTaskID: string;
}

export interface Description {
	MaintenanceMode: string;
	IsQuarantined: boolean;
	Recovering: string;
	AttachedDevice: AttachedDevice;
	Annotation: Annotation;
	Note: Annotation;
	Version: string;
	PowerCycle: boolean;
	PowerCycleState: PowerCycleState;
	LastUpdated: string;
	Battery: number;
	Temperature: { [key: string]: number };
	RunningSwarmingTask: boolean;
	LaunchedSwarming: boolean;
	RecoveryStart: string;
	DeviceUptime: number;
	SSHUserIP: string;
	SuppliedDimensions: SwarmingDimensions;
	Dimensions: SwarmingDimensions;
	TaskRequest?: TaskRequest;
	TaskStarted: string;
}

export type SwarmingDimensions = { [key: string]: string[] | null } | null;

export type AttachedDevice = 'nodevice' | 'adb' | 'ios' | 'ssh';

export type PowerCycleState = 'not_available' | 'available' | 'in_error';

export type Duration = number;

export type ListMachinesResponse = Description[];

export type TaskRequestor = 'swarming' | 'sktask';
