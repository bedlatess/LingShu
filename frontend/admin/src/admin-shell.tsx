import React from "react";
import { AuditOutlined, DashboardOutlined, KeyOutlined, SettingOutlined, TeamOutlined } from "@ant-design/icons";
import { ConfigProvider, Menu } from "antd";
import { useLocation, useNavigate } from "react-router-dom";
import { designTokens } from "@lingshu/shared";

const adminMenuItems = [
  { key: "/dashboard", icon: <DashboardOutlined />, label: "概览" },
  { key: "/users", icon: <TeamOutlined />, label: "用户管理" },
  { key: "/api-keys", icon: <KeyOutlined />, label: "API 密钥" },
  { key: "/models", icon: <SettingOutlined />, label: "模型管理" },
  { key: "/channels", icon: <SettingOutlined />, label: "渠道管理" },
  { key: "/announcements", icon: <SettingOutlined />, label: "公告管理" },
  { key: "/redeem", icon: <KeyOutlined />, label: "兑换码" },
  { key: "/reports", icon: <DashboardOutlined />, label: "数据报表" },
  { key: "/settings", icon: <SettingOutlined />, label: "系统设置" },
  { key: "/audit", icon: <AuditOutlined />, label: "审计日志" }
];

export function AdminMenu() {
  const navigate = useNavigate();
  const location = useLocation();
  const selectedKey = location.pathname.startsWith("/users/") ? "/users" : location.pathname === "/" ? "/dashboard" : location.pathname;
  return <Menu theme="dark" mode="inline" selectedKeys={[selectedKey]} items={adminMenuItems} onClick={({ key }) => navigate(key)} />;
}

export function Theme({ children }: { children: React.ReactNode }) {
  return (
    <ConfigProvider theme={{ token: { colorPrimary: designTokens.colors.brand, borderRadius: 8 } }}>
      {children}
    </ConfigProvider>
  );
}

