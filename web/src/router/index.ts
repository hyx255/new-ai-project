import { createRouter, createWebHistory } from "vue-router";
import type { RouteRecordRaw } from "vue-router";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    redirect: "/dashboard",
  },
  {
    path: "/dashboard",
    name: "Dashboard",
    component: () => import("@/views/DashboardView.vue"),
    meta: { title: "Dashboard" },
  },
  // Device Type routes
  {
    path: "/device-types",
    name: "DeviceTypeList",
    component: () => import("@/views/device-type/DeviceTypeListView.vue"),
    meta: { title: "设备类型" },
  },
  {
    path: "/device-types/create",
    name: "DeviceTypeCreate",
    component: () => import("@/views/device-type/DeviceTypeFormView.vue"),
    meta: { title: "新建设备类型" },
  },
  {
    path: "/device-types/:id/edit",
    name: "DeviceTypeEdit",
    component: () => import("@/views/device-type/DeviceTypeFormView.vue"),
    meta: { title: "编辑设备类型" },
  },
  // Device routes
  {
    path: "/devices",
    name: "DeviceList",
    component: () => import("@/views/device/DeviceListView.vue"),
    meta: { title: "设备管理" },
  },
  {
    path: "/devices/create",
    name: "DeviceCreate",
    component: () => import("@/views/device/DeviceFormView.vue"),
    meta: { title: "新建设备" },
  },
  {
    path: "/devices/:id",
    name: "DeviceDetail",
    component: () => import("@/views/device/DeviceDetailView.vue"),
    meta: { title: "设备详情" },
  },
  {
    path: "/devices/:id/edit",
    name: "DeviceEdit",
    component: () => import("@/views/device/DeviceFormView.vue"),
    meta: { title: "编辑设备" },
  },
  // Batch operation routes
  {
    path: "/batch-operations/:id",
    name: "BatchOperationDetail",
    component: () => import("@/views/device/BatchOperationView.vue"),
    meta: { title: "批量操作" },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const title = (to.meta?.title as string) || "Smart Broadcast Platform";
  document.title = `${title} - Broadcast Platform`;
});

export default router;