import React from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { 
  FiTerminal, 
  FiShield, 
  FiZap, 
  FiCode, 
  FiFileText,
  FiPlay,
  FiLock
} from 'react-icons/fi';

const LandingContainer = styled.div`
  min-height: 100vh;
  background: linear-gradient(135deg, ${props => props.theme.colors.background} 0%, #1a1a1a 100%);
`;

const Hero = styled.section`
  padding: ${props => props.theme.spacing.xxxl} ${props => props.theme.spacing.xl};
  text-align: center;
  max-width: 1200px;
  margin: 0 auto;
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    padding: ${props => props.theme.spacing.xxl} ${props => props.theme.spacing.md};
  }
`;

const HeroTitle = styled(motion.h1)`
  font-size: 4rem;
  background: linear-gradient(135deg, ${props => props.theme.colors.primary} 0%, #8b5cf6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin-bottom: ${props => props.theme.spacing.lg};
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    font-size: 2.5rem;
  }
`;

const HeroSubtitle = styled(motion.p)`
  font-size: 1.5rem;
  color: ${props => props.theme.colors.text.secondary};
  margin-bottom: ${props => props.theme.spacing.xxl};
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    font-size: 1.25rem;
    margin-bottom: ${props => props.theme.spacing.xl};
  }
`;

const CTAButtons = styled(motion.div)`
  display: flex;
  gap: ${props => props.theme.spacing.lg};
  justify-content: center;
  margin-bottom: ${props => props.theme.spacing.xxxl};
  
  @media (max-width: ${props => props.theme.breakpoints.mobile}) {
    flex-direction: column;
    align-items: center;
  }
`;

const Button = styled(Link)`
  display: inline-flex;
  align-items: center;
  gap: ${props => props.theme.spacing.sm};
  padding: ${props => props.theme.spacing.md} ${props => props.theme.spacing.xl};
  border-radius: ${props => props.theme.borderRadius.lg};
  font-weight: 600;
  text-decoration: none;
  transition: all ${props => props.theme.animations.normal};
  
  &.primary {
    background: ${props => props.theme.colors.primary};
    color: white;
    
    &:hover {
      background: ${props => props.theme.colors.primaryHover};
      transform: translateY(-2px);
      box-shadow: ${props => props.theme.shadows.lg};
    }
  }
  
  &.secondary {
    background: transparent;
    color: ${props => props.theme.colors.text.primary};
    border: 2px solid ${props => props.theme.colors.border};
    
    &:hover {
      border-color: ${props => props.theme.colors.primary};
      color: ${props => props.theme.colors.primary};
    }
  }
`;

const SecurityAlert = styled(motion.div)`
  background: linear-gradient(135deg, ${props => props.theme.colors.success}20 0%, ${props => props.theme.colors.info}20 100%);
  border: 1px solid ${props => props.theme.colors.success}40;
  border-radius: ${props => props.theme.borderRadius.lg};
  padding: ${props => props.theme.spacing.lg};
  margin: ${props => props.theme.spacing.xl} auto;
  max-width: 800px;
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.md};
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    flex-direction: column;
    text-align: center;
  }
`;

const Features = styled.section`
  padding: ${props => props.theme.spacing.xxxl} ${props => props.theme.spacing.xl};
  max-width: 1200px;
  margin: 0 auto;
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    padding: ${props => props.theme.spacing.xxl} ${props => props.theme.spacing.md};
  }
`;

const FeaturesGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: ${props => props.theme.spacing.xl};
  margin-top: ${props => props.theme.spacing.xxl};
`;

const FeatureCard = styled(motion.div)`
  background: ${props => props.theme.colors.surface};
  border: 1px solid ${props => props.theme.colors.border};
  border-radius: ${props => props.theme.borderRadius.lg};
  padding: ${props => props.theme.spacing.xl};
  transition: all ${props => props.theme.animations.normal};
  
  &:hover {
    border-color: ${props => props.theme.colors.primary};
    transform: translateY(-4px);
    box-shadow: ${props => props.theme.shadows.lg};
  }
`;

const FeatureIcon = styled.div`
  background: ${props => props.theme.colors.primary}20;
  color: ${props => props.theme.colors.primary};
  width: 60px;
  height: 60px;
  border-radius: ${props => props.theme.borderRadius.lg};
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  margin-bottom: ${props => props.theme.spacing.lg};
`;

const FeatureTitle = styled.h3`
  font-size: 1.25rem;
  margin-bottom: ${props => props.theme.spacing.md};
  color: ${props => props.theme.colors.text.primary};
`;

const FeatureDescription = styled.p`
  color: ${props => props.theme.colors.text.secondary};
  line-height: 1.6;
`;

const SectionTitle = styled.h2`
  font-size: 2.5rem;
  text-align: center;
  margin-bottom: ${props => props.theme.spacing.lg};
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    font-size: 2rem;
  }
`;

const LandingPage = () => {
  const features = [
    {
      icon: FiTerminal,
      title: 'Terminal-Native',
      description: 'Built for developers who live in the terminal. Seamless integration with your existing workflow.'
    },
    {
      icon: FiShield,
      title: 'Security First',
      description: 'Enhanced security with permission controls, path validation, and command sanitization.'
    },
    {
      icon: FiZap,
      title: 'AI-Powered',
      description: 'Leverage multiple AI providers for code analysis, generation, and intelligent assistance.'
    },
    {
      icon: FiCode,
      title: 'LSP Integration',
      description: 'Full Language Server Protocol support for intelligent code understanding and navigation.'
    },
    {
      icon: FiFileText,
      title: 'Context Aware',
      description: 'Automatically understands your project structure and maintains context across conversations.'
    },
    {
      icon: FiLock,
      title: 'Permission System',
      description: 'Granular permission controls ensure AI actions are always under your control.'
    }
  ];

  return (
    <LandingContainer>
      <Hero>
        <HeroTitle
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
        >
          Crush
        </HeroTitle>
        
        <HeroSubtitle
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          Your AI-powered development companion for the terminal
        </HeroSubtitle>
        
        <CTAButtons
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.4 }}
        >
          <Button to="/onboarding" className="primary">
            <FiPlay />
            Get Started
          </Button>
          <Button to="/docs" className="secondary">
            <FiFileText />
            Documentation
          </Button>
        </CTAButtons>
        
        <SecurityAlert
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.6, delay: 0.6 }}
        >
          <FiShield size={24} color="#10b981" />
          <div>
            <strong>Security Enhanced:</strong> This version includes critical security fixes for 
            YOLO mode, command injection, and path traversal vulnerabilities.
          </div>
        </SecurityAlert>
      </Hero>

      <Features>
        <SectionTitle>Why Choose Crush?</SectionTitle>
        
        <FeaturesGrid>
          {features.map((feature, index) => (
            <FeatureCard
              key={feature.title}
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.8 + index * 0.1 }}
            >
              <FeatureIcon>
                <feature.icon />
              </FeatureIcon>
              <FeatureTitle>{feature.title}</FeatureTitle>
              <FeatureDescription>{feature.description}</FeatureDescription>
            </FeatureCard>
          ))}
        </FeaturesGrid>
      </Features>
    </LandingContainer>
  );
};

export default LandingPage;