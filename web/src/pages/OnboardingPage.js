import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  FiArrowLeft,
  FiArrowRight,
  FiTerminal,
  FiShield,
  FiSettings,
  FiCode,
  FiFileText,
  FiZap,
  FiCheck,
  FiAlertTriangle,
  FiLock,
  FiDownload,
  FiPlay
} from 'react-icons/fi';

const OnboardingContainer = styled.div`
  min-height: 100vh;
  padding: ${props => props.theme.spacing.xl};
  background: ${props => props.theme.colors.background};
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    padding: ${props => props.theme.spacing.md};
  }
`;

const OnboardingContent = styled.div`
  max-width: 800px;
  margin: 0 auto;
`;

const Header = styled.div`
  text-align: center;
  margin-bottom: ${props => props.theme.spacing.xxl};
`;

const Title = styled.h1`
  font-size: 2.5rem;
  margin-bottom: ${props => props.theme.spacing.md};
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    font-size: 2rem;
  }
`;

const StepIndicator = styled.div`
  display: flex;
  justify-content: center;
  gap: ${props => props.theme.spacing.sm};
  margin-bottom: ${props => props.theme.spacing.xl};
`;

const StepDot = styled.div`
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: ${props => props.isActive ? props.theme.colors.primary : props.theme.colors.border};
  transition: all ${props => props.theme.animations.normal};
`;

const StepContent = styled(motion.div)`
  background: ${props => props.theme.colors.surface};
  border: 1px solid ${props => props.theme.colors.border};
  border-radius: ${props => props.theme.borderRadius.lg};
  padding: ${props => props.theme.spacing.xxl};
  margin-bottom: ${props => props.theme.spacing.xl};
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    padding: ${props => props.theme.spacing.xl};
  }
`;

const StepTitle = styled.h2`
  font-size: 1.5rem;
  margin-bottom: ${props => props.theme.spacing.lg};
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.md};
`;

const ToolGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: ${props => props.theme.spacing.lg};
  margin: ${props => props.theme.spacing.lg} 0;
`;

const ToolCard = styled.div`
  background: ${props => props.theme.colors.background};
  border: 1px solid ${props => props.theme.colors.border};
  border-radius: ${props => props.theme.borderRadius.md};
  padding: ${props => props.theme.spacing.lg};
  transition: all ${props => props.theme.animations.normal};
  
  &:hover {
    border-color: ${props => props.theme.colors.primary};
  }
`;

const ToolIcon = styled.div`
  color: ${props => props.theme.colors.primary};
  font-size: 1.5rem;
  margin-bottom: ${props => props.theme.spacing.sm};
`;

const ToolName = styled.h4`
  font-size: 1rem;
  margin-bottom: ${props => props.theme.spacing.sm};
`;

const ToolDescription = styled.p`
  color: ${props => props.theme.colors.text.secondary};
  font-size: 0.875rem;
  line-height: 1.5;
`;

const SecurityWarning = styled.div`
  background: ${props => props.theme.colors.warning}20;
  border: 1px solid ${props => props.theme.colors.warning}40;
  border-radius: ${props => props.theme.borderRadius.md};
  padding: ${props => props.theme.spacing.lg};
  margin: ${props => props.theme.spacing.lg} 0;
  display: flex;
  gap: ${props => props.theme.spacing.md};
  align-items: flex-start;
`;

const CodeBlock = styled.pre`
  background: ${props => props.theme.colors.terminal.background};
  color: ${props => props.theme.colors.terminal.text};
  border-radius: ${props => props.theme.borderRadius.md};
  padding: ${props => props.theme.spacing.lg};
  font-family: ${props => props.theme.fonts.mono};
  font-size: 0.875rem;
  overflow-x: auto;
  margin: ${props => props.theme.spacing.lg} 0;
`;

const Navigation = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

const Button = styled.button`
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.sm};
  padding: ${props => props.theme.spacing.md} ${props => props.theme.spacing.lg};
  border-radius: ${props => props.theme.borderRadius.md};
  font-weight: 500;
  transition: all ${props => props.theme.animations.normal};
  
  &.primary {
    background: ${props => props.theme.colors.primary};
    color: white;
    
    &:hover {
      background: ${props => props.theme.colors.primaryHover};
    }
  }
  
  &.secondary {
    background: transparent;
    color: ${props => props.theme.colors.text.secondary};
    
    &:hover {
      color: ${props => props.theme.colors.text.primary};
    }
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

const LinkButton = styled(Link)`
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.sm};
  padding: ${props => props.theme.spacing.md} ${props => props.theme.spacing.lg};
  border-radius: ${props => props.theme.borderRadius.md};
  font-weight: 500;
  text-decoration: none;
  transition: all ${props => props.theme.animations.normal};
  background: ${props => props.theme.colors.primary};
  color: white;
  
  &:hover {
    background: ${props => props.theme.colors.primaryHover};
  }
`;

const OnboardingPage = () => {
  const [currentStep, setCurrentStep] = useState(0);

  const steps = [
    {
      title: 'Welcome to Crush',
      icon: FiTerminal,
      content: (
        <>
          <p>
            Crush is your AI-powered development companion that lives right in your terminal. 
            It helps you write code, analyze projects, and automate development tasks with the power of AI.
          </p>
          
          <ToolGrid>
            <ToolCard>
              <ToolIcon><FiCode /></ToolIcon>
              <ToolName>Code Analysis</ToolName>
              <ToolDescription>
                Understand codebases, analyze patterns, and get intelligent insights about your projects.
              </ToolDescription>
            </ToolCard>
            <ToolCard>
              <ToolIcon><FiFileText /></ToolIcon>
              <ToolName>File Operations</ToolName>
              <ToolDescription>
                Read, write, edit, and search through files with AI assistance and safety checks.
              </ToolDescription>
            </ToolCard>
            <ToolCard>
              <ToolIcon><FiTerminal /></ToolIcon>
              <ToolName>Shell Commands</ToolName>
              <ToolDescription>
                Execute shell commands safely with permission controls and security validation.
              </ToolDescription>
            </ToolCard>
            <ToolCard>
              <ToolIcon><FiZap /></ToolIcon>
              <ToolName>AI Integration</ToolName>
              <ToolDescription>
                Support for multiple AI providers: OpenAI, Anthropic, Google, and more.
              </ToolDescription>
            </ToolCard>
          </ToolGrid>
        </>
      )
    },
    {
      title: 'Security Features',
      icon: FiShield,
      content: (
        <>
          <p>
            Security is our top priority. Crush includes multiple layers of protection to keep your system safe:
          </p>
          
          <ToolGrid>
            <ToolCard>
              <ToolIcon><FiLock /></ToolIcon>
              <ToolName>Permission System</ToolName>
              <ToolDescription>
                Every operation requires explicit permission. You control what the AI can access and execute.
              </ToolDescription>
            </ToolCard>
            <ToolCard>
              <ToolIcon><FiShield /></ToolIcon>
              <ToolName>Path Validation</ToolName>
              <ToolDescription>
                All file operations are validated to prevent directory traversal and unauthorized access.
              </ToolDescription>
            </ToolCard>
            <ToolCard>
              <ToolIcon><FiCheck /></ToolIcon>
              <ToolName>Command Sanitization</ToolName>
              <ToolDescription>
                Shell commands are validated against allowlists and dangerous patterns are blocked.
              </ToolDescription>
            </ToolCard>
          </ToolGrid>

          <SecurityWarning>
            <FiAlertTriangle color="#f59e0b" />
            <div>
              <strong>YOLO Mode Warning:</strong> YOLO mode bypasses all security checks. 
              Only use it in isolated testing environments where system security is not a concern.
            </div>
          </SecurityWarning>
        </>
      )
    },
    {
      title: 'Installation & Setup',
      icon: FiDownload,
      content: (
        <>
          <p>
            Get started with Crush in just a few steps:
          </p>
          
          <h4>1. Download and Install</h4>
          <CodeBlock>
{`# Download the latest release
curl -sSL https://github.com/ashusingh22/crush/releases/latest/download/crush-linux-amd64 -o crush

# Make it executable
chmod +x crush

# Move to PATH
sudo mv crush /usr/local/bin/`}
          </CodeBlock>

          <h4>2. Initial Configuration</h4>
          <CodeBlock>
{`# Run Crush for the first time
crush

# Or start with an API key
crush --help`}
          </CodeBlock>

          <h4>3. Set up AI Provider</h4>
          <p>
            Crush supports multiple AI providers. You'll need an API key from at least one:
          </p>
          <ul>
            <li>OpenAI (GPT-4, GPT-3.5)</li>
            <li>Anthropic (Claude)</li>
            <li>Google (Gemini)</li>
            <li>Custom providers</li>
          </ul>
        </>
      )
    },
    {
      title: 'Core Tools & Features',
      icon: FiSettings,
      content: (
        <>
          <p>
            Crush provides a comprehensive set of tools for development workflows:
          </p>
          
          <ToolGrid>
            <ToolCard>
              <ToolIcon><FiFileText /></ToolIcon>
              <ToolName>view</ToolName>
              <ToolDescription>
                Read and display file contents with syntax highlighting and context awareness.
              </ToolDescription>
            </ToolCard>
            <ToolCard>
              <ToolIcon><FiCode /></ToolIcon>
              <ToolName>edit</ToolName>
              <ToolDescription>
                Make targeted edits to files with find-and-replace or append operations.
              </ToolDescription>
            </ToolCard>
            <ToolCard>
              <ToolIcon><FiFileText /></ToolIcon>
              <ToolName>write</ToolName>
              <ToolDescription>
                Create new files or completely rewrite existing ones with AI assistance.
              </ToolDescription>
            </ToolCard>
            <ToolCard>
              <ToolIcon><FiTerminal /></ToolIcon>
              <ToolName>bash</ToolName>
              <ToolDescription>
                Execute shell commands safely with permission controls and validation.
              </ToolDescription>
            </ToolCard>
            <ToolCard>
              <ToolIcon><FiDownload /></ToolIcon>
              <ToolName>download</ToolName>
              <ToolDescription>
                Download files from URLs with security validation and path checks.
              </ToolDescription>
            </ToolCard>
            <ToolCard>
              <ToolIcon><FiFileText /></ToolIcon>
              <ToolName>grep</ToolName>
              <ToolDescription>
                Search for patterns in files and directories with powerful filtering.
              </ToolDescription>
            </ToolCard>
          </ToolGrid>
        </>
      )
    },
    {
      title: 'Ready to Start!',
      icon: FiPlay,
      content: (
        <>
          <p>
            You're all set! Here are some example commands to get you started:
          </p>
          
          <CodeBlock>
{`# Start an interactive session
crush

# Analyze your current project
crush run "Analyze this codebase and explain its structure"

# Get help with debugging
crush run "Help me debug this error: [paste error here]"

# Generate documentation
crush run "Create a README for this project"`}
          </CodeBlock>

          <SecurityWarning>
            <FiShield color="#10b981" />
            <div>
              <strong>Security Tip:</strong> Always review permissions before granting them. 
              Crush will ask for your approval before making any changes to your system.
            </div>
          </SecurityWarning>

          <p>
            Ready to experience AI-powered development? Start your first chat session!
          </p>
        </>
      )
    }
  ];

  const nextStep = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const prevStep = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const currentStepData = steps[currentStep];

  return (
    <OnboardingContainer>
      <OnboardingContent>
        <Header>
          <Title>Get Started with Crush</Title>
          <StepIndicator>
            {steps.map((_, index) => (
              <StepDot key={index} isActive={index <= currentStep} />
            ))}
          </StepIndicator>
        </Header>

        <AnimatePresence mode="wait">
          <StepContent
            key={currentStep}
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -20 }}
            transition={{ duration: 0.3 }}
          >
            <StepTitle>
              <currentStepData.icon />
              {currentStepData.title}
            </StepTitle>
            {currentStepData.content}
          </StepContent>
        </AnimatePresence>

        <Navigation>
          <Button 
            className="secondary" 
            onClick={prevStep} 
            disabled={currentStep === 0}
          >
            <FiArrowLeft />
            Previous
          </Button>

          {currentStep === steps.length - 1 ? (
            <LinkButton to="/chat">
              <FiPlay />
              Start Chatting
            </LinkButton>
          ) : (
            <Button className="primary" onClick={nextStep}>
              Next
              <FiArrowRight />
            </Button>
          )}
        </Navigation>
      </OnboardingContent>
    </OnboardingContainer>
  );
};

export default OnboardingPage;