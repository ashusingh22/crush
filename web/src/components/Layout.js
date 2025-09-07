import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { 
  FiHome, 
  FiMessageCircle, 
  FiBook, 
  FiMenu, 
  FiX, 
  FiShield,
  FiTerminal 
} from 'react-icons/fi';

const LayoutContainer = styled.div`
  display: flex;
  min-height: 100vh;
  background: ${props => props.theme.colors.background};
`;

const Sidebar = styled(motion.nav)`
  width: 280px;
  background: ${props => props.theme.colors.surface};
  border-right: 1px solid ${props => props.theme.colors.border};
  display: flex;
  flex-direction: column;
  position: fixed;
  height: 100vh;
  z-index: 100;
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    width: 100%;
    transform: translateX(-100%);
    ${props => props.isOpen && `transform: translateX(0);`}
  }
`;

const SidebarHeader = styled.div`
  padding: ${props => props.theme.spacing.lg};
  border-bottom: 1px solid ${props => props.theme.colors.border};
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.md};
`;

const Logo = styled.div`
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.sm};
  font-size: 1.5rem;
  font-weight: 700;
  color: ${props => props.theme.colors.primary};
`;

const NavList = styled.ul`
  list-style: none;
  padding: ${props => props.theme.spacing.md};
  flex: 1;
`;

const NavItem = styled.li`
  margin-bottom: ${props => props.theme.spacing.sm};
`;

const NavLink = styled(Link)`
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.md};
  padding: ${props => props.theme.spacing.md};
  border-radius: ${props => props.theme.borderRadius.md};
  color: ${props => props.theme.colors.text.secondary};
  text-decoration: none;
  transition: all ${props => props.theme.animations.normal};
  
  ${props => props.isActive && `
    background: ${props.theme.colors.primary};
    color: white;
  `}
  
  &:hover {
    background: ${props => props.theme.colors.surfaceHover};
    color: ${props => props.theme.colors.text.primary};
  }
`;

const MainContent = styled.main`
  flex: 1;
  margin-left: 280px;
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    margin-left: 0;
  }
`;

const Header = styled.header`
  display: none;
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: ${props => props.theme.spacing.md};
    background: ${props => props.theme.colors.surface};
    border-bottom: 1px solid ${props => props.theme.colors.border};
  }
`;

const MenuButton = styled.button`
  background: none;
  color: ${props => props.theme.colors.text.primary};
  font-size: 1.5rem;
  padding: ${props => props.theme.spacing.sm};
`;

const SecurityBadge = styled.div`
  margin-top: auto;
  padding: ${props => props.theme.spacing.md};
  background: ${props => props.theme.colors.success}20;
  border: 1px solid ${props => props.theme.colors.success}40;
  border-radius: ${props => props.theme.borderRadius.md};
  margin: ${props => props.theme.spacing.md};
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.sm};
  color: ${props => props.theme.colors.success};
  font-size: 0.875rem;
`;

const Overlay = styled(motion.div)`
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 50;
  
  @media (min-width: ${props => props.theme.breakpoints.tablet}) {
    display: none;
  }
`;

const Layout = ({ children }) => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const location = useLocation();

  const navItems = [
    { path: '/', label: 'Home', icon: FiHome },
    { path: '/onboarding', label: 'Get Started', icon: FiBook },
    { path: '/chat', label: 'Chat', icon: FiMessageCircle },
    { path: '/docs', label: 'Documentation', icon: FiBook },
  ];

  const closeSidebar = () => setSidebarOpen(false);

  return (
    <LayoutContainer>
      {sidebarOpen && (
        <Overlay
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={closeSidebar}
        />
      )}
      
      <Sidebar
        isOpen={sidebarOpen}
        initial={false}
        animate={{
          x: sidebarOpen ? 0 : '-100%',
        }}
        transition={{ type: 'spring', stiffness: 300, damping: 30 }}
      >
        <SidebarHeader>
          <Logo>
            <FiTerminal />
            Crush
          </Logo>
          <MenuButton 
            onClick={closeSidebar}
            style={{ marginLeft: 'auto', display: sidebarOpen ? 'block' : 'none' }}
          >
            <FiX />
          </MenuButton>
        </SidebarHeader>
        
        <NavList>
          {navItems.map(({ path, label, icon: Icon }) => (
            <NavItem key={path}>
              <NavLink 
                to={path} 
                isActive={location.pathname === path}
                onClick={closeSidebar}
              >
                <Icon />
                {label}
              </NavLink>
            </NavItem>
          ))}
        </NavList>
        
        <SecurityBadge>
          <FiShield />
          Security Enhanced
        </SecurityBadge>
      </Sidebar>

      <MainContent>
        <Header>
          <MenuButton onClick={() => setSidebarOpen(true)}>
            <FiMenu />
          </MenuButton>
          <Logo>
            <FiTerminal />
            Crush
          </Logo>
        </Header>
        {children}
      </MainContent>
    </LayoutContainer>
  );
};

export default Layout;