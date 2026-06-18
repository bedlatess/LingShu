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
  return <Menu className="admin-menu" theme="light" mode="inline" selectedKeys={[selectedKey]} items={adminMenuItems} onClick={({ key }) => navigate(key)} />;
}

export function Theme({ children }: { children: React.ReactNode }) {
  return (
    <ConfigProvider
      theme={{
        cssVar: { key: "lingshu-admin" },
        token: {
          colorPrimary: designTokens.colors.clay,
          colorInfo: designTokens.colors.clay,
          colorSuccess: designTokens.colors.success,
          colorWarning: designTokens.colors.warning,
          colorError: designTokens.colors.danger,
          colorBgLayout: designTokens.colors.bg,
          colorBgContainer: designTokens.colors.surface,
          colorBgElevated: designTokens.colors.surface,
          colorBorder: designTokens.colors.border,
          colorBorderSecondary: designTokens.colors.border,
          colorText: designTokens.colors.ink,
          colorTextSecondary: designTokens.colors.inkMuted,
          borderRadius: 6,
          fontFamily: designTokens.font.sans,
          boxShadow: designTokens.shadow.md,
          boxShadowSecondary: designTokens.shadow.sm
        },
        components: {
          Layout: {
            headerBg: designTokens.colors.surface,
            siderBg: designTokens.colors.bgSubtle,
            bodyBg: designTokens.colors.bg
          },
          Menu: {
            itemBg: "transparent",
            itemSelectedBg: designTokens.colors.claySoft,
            itemSelectedColor: designTokens.colors.clayHover,
            itemColor: designTokens.colors.inkMuted
          },
          Table: { headerBg: designTokens.colors.bgSubtle, headerColor: designTokens.colors.ink },
          Card: { colorBorderSecondary: designTokens.colors.border }
        }
      }}
    >
      {children}
    </ConfigProvider>
  );
}
