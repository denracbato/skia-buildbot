import { ListMachinesResponse } from '../json';

export const fakeNow = Date.parse('2021-06-03T18:20:30.00000Z');

// Based on a production response on 2021-06-03.
export const descriptions: ListMachinesResponse = [{
  Mode: 'available',
  AttachedDevice: 'adb',
  Annotation: {
    Message: 'Leaving recovery mode.',
    User: 'machines.skia.org',
    Timestamp: '2021-06-03T18:20:24.97453Z',
  },
  Note: {
    Message: '',
    User: '',
    Timestamp: '0001-01-01T00:00:00Z',
  },
  Dimensions: {
    android_devices: ['1'],
    device_os: ['Q', 'QP1A.190711.020', 'QP1A.190711.020_G980FXXU1ATB3'],
    device_os_flavor: ['samsung'],
    device_os_type: ['user'],
    device_type: ['x1s', 'exynos990'],
    id: ['skia-rpi2-rack4-shelf1-001'],
    inside_docker: ['1', 'containerd'],
    os: ['Android'],
  },
  Version: 'v1.2',
  PowerCycle: false,
  PowerCycleState: 'available',
  LastUpdated: '2021-06-03T18:20:24.974527Z',
  Battery: 100,
  Temperature: {
    TYPE_BATTERY: 24.2,
    TYPE_CPU: 29.1,
    TYPE_SKIN: 26.3,
    TYPE_USB_PORT: 23.2,
    dumpsys_battery: 24.3,
  },
  RunningSwarmingTask: true,
  LaunchedSwarming: false,
  DeviceUptime: 167,
  SSHUserIP: '',
}, {
  Mode: 'available',
  AttachedDevice: 'ssh',
  Annotation: {
    Message: 'Pod too old, requested update for "rpi-swarming-qdgf2"',
    User: 'machines.skia.org',
    Timestamp: '2021-06-03T18:20:18.710419Z',
  },
  Note: {
    Message: '',
    User: '',
    Timestamp: '0001-01-01T00:00:00Z',
  },
  Dimensions: {
    android_devices: ['1'],
    device_os: ['H', 'HUAWEIELE-L29', 'HUAWEIELE-L29_9.1.0.241C605'],
    device_os_flavor: ['huawei'],
    device_os_type: ['user'],
    device_type: ['HWELE', 'ELE'],
    id: ['skia-rpi2-rack4-shelf1-002'],
    inside_docker: ['1', 'containerd'],
    os: ['Android'],
  },
  Version: 'v1.3',
  PowerCycle: false,
  PowerCycleState: 'in_error',
  LastUpdated: '2021-06-03T18:20:18.710416Z',
  Battery: 100,
  Temperature: {
    dumpsys_battery: 23,
  },
  RunningSwarmingTask: true,
  LaunchedSwarming: true,
  DeviceUptime: 266,
  SSHUserIP: '',
}, {
  Mode: 'available',
  AttachedDevice: 'adb',
  Annotation: {
    Message: 'Pod too old, requested update for "rpi-swarming-5hqvb"',
    User: 'machines.skia.org',
    Timestamp: '2021-06-03T18:20:18.967714Z',
  },
  Note: {
    Message: '',
    User: '',
    Timestamp: '0001-01-01T00:00:00Z',
  },
  Dimensions: {
    android_devices: ['1'],
    device_os: ['H', 'HUAWEIELE-L29', 'HUAWEIELE-L29_9.1.0.241C605'],
    device_os_flavor: ['huawei'],
    device_os_type: ['user'],
    device_type: ['HWELE', 'ELE'],
    id: ['skia-rpi2-rack4-shelf1-003'],
    inside_docker: ['1', 'containerd'],
    os: ['Android'],
  },
  Version: 'v1.4',
  PowerCycle: false,
  PowerCycleState: 'not_available',
  LastUpdated: '2021-06-03T18:20:20.87764Z',
  Battery: 100,
  Temperature: {
    dumpsys_battery: 22,
  },
  RunningSwarmingTask: true,
  LaunchedSwarming: true,
  DeviceUptime: 183,
  SSHUserIP: '',
}, {
  Mode: 'available',
  AttachedDevice: 'adb',
  Annotation: {
    Message: 'Pod too old, requested update for "rpi-swarming-q2vpj"',
    User: 'machines.skia.org',
    Timestamp: '2021-06-03T18:15:13.910199Z',
  },
  Note: {
    Message: '',
    User: '',
    Timestamp: '0001-01-01T00:00:00Z',
  },
  Dimensions: {
    android_devices: ['1'],
    device_os: ['H', 'HUAWEIELE-L29', 'HUAWEIELE-L29_9.1.0.241C605'],
    device_os_flavor: ['huawei'],
    device_os_type: ['user'],
    device_type: ['HWELE', 'ELE'],
    id: ['skia-rpi2-rack4-shelf1-004'],
    inside_docker: ['1', 'containerd'],
    os: ['Android'],
  },
  Version: 'v1.2',
  PowerCycle: false,
  PowerCycleState: 'available',
  LastUpdated: '2021-06-03T18:20:02.034149Z',
  Battery: 100,
  Temperature: {
    dumpsys_battery: 22,
  },
  RunningSwarmingTask: true,
  LaunchedSwarming: true,
  DeviceUptime: 167,
  SSHUserIP: '',
}, {
  Mode: 'available',
  AttachedDevice: 'adb',
  Annotation: {
    Message: 'Pod too old, requested update for "rpi-swarming-k8fdn"',
    User: 'machines.skia.org',
    Timestamp: '2021-06-03T18:19:56.440311Z',
  },
  Note: {
    Message: '',
    User: '',
    Timestamp: '0001-01-01T00:00:00Z',
  },
  Dimensions: {
    android_devices: ['1'],
    device_os: ['H', 'HUAWEIELE-L29', 'HUAWEIELE-L29_9.1.0.241C605'],
    device_os_flavor: ['huawei'],
    device_os_type: ['user'],
    device_type: ['HWELE', 'ELE'],
    id: ['skia-rpi2-rack4-shelf1-005'],
    inside_docker: ['1', 'containerd'],
    os: ['Android'],
  },
  Version: 'v1.2',
  PowerCycle: false,
  PowerCycleState: 'available',
  LastUpdated: '2021-06-03T18:20:21.471856Z',
  Battery: 100,
  Temperature: {
    dumpsys_battery: 23,
  },
  RunningSwarmingTask: true,
  LaunchedSwarming: true,
  DeviceUptime: 234,
  SSHUserIP: '',
}, {
  Mode: 'recovery',
  AttachedDevice: 'adb',
  Annotation: {
    Message: 'Pod too old, requested update for "rpi-swarming-j9lzl"',
    User: 'machines.skia.org',
    Timestamp: '2021-06-03T18:20:13.827511Z',
  },
  Note: {
    Message: '',
    User: '',
    Timestamp: '0001-01-01T00:00:00Z',
  },
  Dimensions: {
    android_devices: ['1'],
    device_os: ['Q', 'QP1A.190711.020', 'QP1A.190711.020_G980FXXU1ATBM'],
    device_os_flavor: ['samsung'],
    device_os_type: ['user'],
    device_type: ['x1s', 'exynos990'],
    id: ['skia-rpi2-rack4-shelf1-006'],
    inside_docker: ['1', 'containerd'],
    os: ['Android'],
  },
  Version: 'v1.2',
  PowerCycle: false,
  PowerCycleState: 'available',
  LastUpdated: '2021-06-03T18:20:33.121421Z',
  Battery: 94,
  Temperature: {
    TYPE_BATTERY: 24.9,
    TYPE_CPU: 36.2,
    TYPE_SKIN: 28.8,
    TYPE_USB_PORT: 23.8,
    dumpsys_battery: 24.9,
  },
  RunningSwarmingTask: true,
  LaunchedSwarming: true,
  DeviceUptime: 343,
  SSHUserIP: '',
}, {
  Mode: 'available',
  AttachedDevice: 'adb',
  Annotation: {
    Message: 'Pod too old, requested update for "rpi-swarming-d86nk"',
    User: 'machines.skia.org',
    Timestamp: '2021-06-03T18:19:49.07976Z',
  },
  Note: {
    Message: '',
    User: '',
    Timestamp: '0001-01-01T00:00:00Z',
  },
  Dimensions: {
    android_devices: ['1'],
    device_os: ['P', 'PQ1A.190105.004', 'PQ1A.190105.004_5148680'],
    device_os_flavor: ['google'],
    device_os_type: ['userdebug'],
    device_type: ['blueline'],
    id: ['skia-rpi2-rack4-shelf1-007'],
    inside_docker: ['1', 'containerd'],
    os: ['Android'],
  },
  Version: 'v1.2',
  PowerCycle: false,
  PowerCycleState: 'available',
  LastUpdated: '2021-06-03T18:20:09.386348Z',
  Battery: 99,
  Temperature: {
    dumpsys_battery: 24.9,
  },
  RunningSwarmingTask: true,
  LaunchedSwarming: true,
  DeviceUptime: 657,
  SSHUserIP: '',
}, {
  Mode: 'available',
  AttachedDevice: 'adb',
  Annotation: {
    Message: 'Leaving recovery mode.',
    User: 'machines.skia.org',
    Timestamp: '2021-06-03T18:14:53.393161Z',
  },
  Note: {
    Message: '',
    User: '',
    Timestamp: '0001-01-01T00:00:00Z',
  },
  Dimensions: {
    android_devices: ['1'],
    device_os: ['Q', 'QP1A.190711.020', 'QP1A.190711.020_G980FXXU1ATB3'],
    device_os_flavor: ['samsung'],
    device_os_type: ['user'],
    device_type: ['x1s', 'exynos990'],
    id: ['skia-rpi2-rack4-shelf1-008'],
    inside_docker: ['1', 'containerd'],
    os: ['Android'],
  },
  Version: 'v1.2',
  PowerCycle: false,
  PowerCycleState: 'available',
  LastUpdated: '2021-06-03T18:20:35.74389Z',
  Battery: 100,
  Temperature: {
    TYPE_BATTERY: 24.2,
    TYPE_CPU: 25.2,
    TYPE_SKIN: 25.5,
    TYPE_USB_PORT: 24,
    dumpsys_battery: 24.2,
  },
  RunningSwarmingTask: true,
  LaunchedSwarming: true,
  DeviceUptime: 660,
  SSHUserIP: '',
}, {
  Mode: 'recovery',
  AttachedDevice: 'adb',
  Annotation: {
    Message: 'Too hot. ',
    User: 'machines.skia.org',
    Timestamp: '2021-06-03T18:20:05.427786Z',
  },
  Note: {
    Message: '',
    User: '',
    Timestamp: '0001-01-01T00:00:00Z',
  },
  Dimensions: {
    android_devices: ['1'],
    device_os: ['Q', 'QP1A.190711.020', 'QP1A.190711.020_G980FXXU1ATB3'],
    device_os_flavor: ['samsung'],
    device_os_type: ['user'],
    device_type: ['x1s', 'exynos990'],
    id: ['skia-rpi2-rack4-shelf1-009'],
    inside_docker: ['1', 'containerd'],
    os: ['Android'],
  },
  Version: 'v1.2',
  PowerCycle: false,
  PowerCycleState: 'available',
  LastUpdated: '2021-06-03T18:20:35.282608Z',
  Battery: 91,
  Temperature: {
    TYPE_BATTERY: 28.8,
    TYPE_CPU: 41.4,
    TYPE_SKIN: 32.3,
    TYPE_USB_PORT: 26.4,
    dumpsys_battery: 28.8,
  },
  RunningSwarmingTask: true,
  LaunchedSwarming: true,
  DeviceUptime: 891,
  SSHUserIP: '',
}, {
  Mode: 'available',
  AttachedDevice: 'ssh',
  Annotation: {
    Message: 'Pod too old, requested update for "rpi-swarming-ghncz"',
    User: 'machines.skia.org',
    Timestamp: '2021-06-03T18:19:55.480856Z',
  },
  Note: {
    Message: '',
    User: '',
    Timestamp: '0001-01-01T00:00:00Z',
  },
  Dimensions: {
    android_devices: ['1'],
    device_os: ['Q', 'QP1A.190711.020', 'QP1A.190711.020_G980FXXU1ATBM'],
    device_os_flavor: ['samsung'],
    device_os_type: ['user'],
    device_type: ['x1s', 'exynos990'],
    id: ['skia-rpi2-rack4-shelf1-010'],
    inside_docker: ['1', 'containerd'],
    os: ['Android'],
  },
  Version: 'v1.2',
  PowerCycle: false,
  PowerCycleState: 'available',
  LastUpdated: '2021-06-03T18:20:23.312632Z',
  Battery: 100,
  Temperature: {
    TYPE_BATTERY: 25.1,
    TYPE_CPU: 26.2,
    TYPE_SKIN: 26.5,
    TYPE_USB_PORT: 25,
    dumpsys_battery: 25,
  },
  RunningSwarmingTask: false,
  LaunchedSwarming: true,
  DeviceUptime: 899,
  SSHUserIP: '',
}];
