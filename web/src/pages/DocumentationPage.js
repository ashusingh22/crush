import React from 'react';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { 
  FiShield, 
  FiTerminal, 
  FiCode, 
  FiFileText,
  FiSettings,
  FiAlertTriangle,
  FiLock,
  FiDownload,
  FiGithub
} from 'react-icons/fi';

const DocsContainer = styled.div`
  max-width: 1000px;
  margin: 0 auto;
  padding: ${props => props.theme.spacing.xl};
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    padding: ${props => props.theme.spacing.md};
  }
`;

const Header = styled.div`
  text-align: center;
  margin-bottom: ${props => props.theme.spacing.xxxl};
`;

const Title = styled.h1`
  font-size: 3rem;
  margin-bottom: ${props => props.theme.spacing.md};
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    font-size: 2rem;
  }
`;

const Subtitle = styled.p`
  font-size: 1.25rem;
  color: ${props => props.theme.colors.text.secondary};
`;

const Section = styled(motion.section)`
  margin-bottom: ${props => props.theme.spacing.xxxl};
`;

const SectionTitle = styled.h2`
  font-size: 2rem;
  margin-bottom: ${props => props.theme.spacing.lg};
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.md};
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    font-size: 1.5rem;
  }
`;

const SecurityAlert = styled.div`
  background: ${props => props.theme.colors.warning}20;
  border: 1px solid ${props => props.theme.colors.warning}40;
  border-radius: ${props => props.theme.borderRadius.lg};
  padding: ${props => props.theme.spacing.lg};
  margin: ${props => props.theme.spacing.lg} 0;
  display: flex;
  gap: ${props => props.theme.spacing.md};
  align-items: flex-start;
`;

const SuccessAlert = styled.div`
  background: ${props => props.theme.colors.success}20;
  border: 1px solid ${props => props.theme.colors.success}40;
  border-radius: ${props => props.theme.borderRadius.lg};
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
  border: 1px solid ${props => props.theme.colors.border};
`;

const Grid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: ${props => props.theme.spacing.lg};
  margin: ${props => props.theme.spacing.lg} 0;
`;

const Card = styled.div`
  background: ${props => props.theme.colors.surface};
  border: 1px solid ${props => props.theme.colors.border};
  border-radius: ${props => props.theme.borderRadius.lg};
  padding: ${props => props.theme.spacing.lg};
  transition: all ${props => props.theme.animations.normal};
  
  &:hover {
    border-color: ${props => props.theme.colors.primary};
  }
`;

const CardIcon = styled.div`
  color: ${props => props.theme.colors.primary};
  font-size: 1.5rem;
  margin-bottom: ${props => props.theme.spacing.md};
`;

const CardTitle = styled.h3`
  margin-bottom: ${props => props.theme.spacing.sm};
`;

const CardDescription = styled.p`
  color: ${props => props.theme.colors.text.secondary};
  line-height: 1.6;
`;

const List = styled.ul`
  margin: ${props => props.theme.spacing.md} 0;
  padding-left: ${props => props.theme.spacing.lg};
  
  li {
    margin-bottom: ${props => props.theme.spacing.sm};
    line-height: 1.6;
  }
`;

const DocumentationPage = () => {
  return (
    <DocsContainer>
      <Header>
        <Title>Crush Documentation</Title>
        <Subtitle>Your complete guide to AI-powered development</Subtitle>
      </Header>

      <Section
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
      >
        <SectionTitle>
          <FiShield />
          Security Enhancements
        </SectionTitle>
        
        <p>
          This version of Crush includes critical security fixes that make it safe for production use:
        </p>

        <SuccessAlert>
          <FiLock color="#10b981" />
          <div>
            <strong>Security Fixes Implemented:</strong>
            <List>
              <li>YOLO mode now requires explicit confirmation and provides warnings</li>
              <li>Command injection prevention with allowlists and pattern detection</li>
              <li>Path traversal protection for all file operations</li>
              <li>Enhanced audit logging for security events</li>
            </List>
          </div>
        </SuccessAlert>

        <SecurityAlert>
          <FiAlertTriangle color="#f59e0b" />
          <div>
            <strong>YOLO Mode Warning:</strong> The <code>--yolo</code> flag completely disables all security 
            measures. Only use this in completely isolated testing environments where system security is not a concern.
          </div>
        </SecurityAlert>
      </Section>

      <Section
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.2 }}
      >
        <SectionTitle>
          <FiDownload />
          Installation
        </SectionTitle>
        
        <p>Install Crush on your system with these simple steps:</p>

        <h3>Download Binary</h3>
        <CodeBlock>
{`# Linux/macOS
curl -sSL https://github.com/ashusingh22/crush/releases/latest/download/crush-linux-amd64 -o crush
chmod +x crush
sudo mv crush /usr/local/bin/

# Windows
# Download from GitHub releases page`}
        </CodeBlock>

        <h3>Build from Source</h3>
        <CodeBlock>
{`git clone https://github.com/ashusingh22/crush.git
cd crush
go build -o crush .
sudo mv crush /usr/local/bin/`}
        </CodeBlock>
      </Section>

      <Section
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.4 }}
      >
        <SectionTitle>
          <FiTerminal />
          Core Tools
        </SectionTitle>
        
        <p>Crush provides a comprehensive set of AI-powered development tools:</p>

        <Grid>
          <Card>
            <CardIcon><FiFileText /></CardIcon>
            <CardTitle>view</CardTitle>
            <CardDescription>
              Read and display file contents with intelligent context awareness and syntax highlighting.
            </CardDescription>
          </Card>
          
          <Card>
            <CardIcon><FiCode /></CardIcon>
            <CardTitle>edit</CardTitle>
            <CardDescription>
              Make precise edits to files using find-and-replace or targeted modifications.
            </CardDescription>
          </Card>
          
          <Card>
            <CardIcon><FiFileText /></CardIcon>
            <CardTitle>write</CardTitle>
            <CardDescription>
              Create new files or completely rewrite existing ones with AI assistance.
            </CardDescription>
          </Card>
          
          <Card>
            <CardIcon><FiTerminal /></CardIcon>
            <CardTitle>bash</CardTitle>
            <CardDescription>
              Execute shell commands safely with permission controls and security validation.
            </CardDescription>
          </Card>
          
          <Card>
            <CardIcon><FiDownload /></CardIcon>
            <CardTitle>download</CardTitle>
            <CardDescription>
              Download files from URLs with automatic security checks and path validation.
            </CardDescription>
          </Card>
          
          <Card>
            <CardIcon><FiFileText /></CardIcon>
            <CardTitle>grep</CardTitle>
            <CardDescription>
              Search for patterns in files and directories with advanced filtering options.
            </CardDescription>
          </Card>
        </Grid>
      </Section>

      <Section
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.6 }}
      >
        <SectionTitle>
          <FiSettings />
          Configuration
        </SectionTitle>
        
        <p>Configure Crush for your development workflow:</p>

        <h3>Basic Configuration</h3>
        <CodeBlock>
{`# Start Crush interactively
crush

# Run with debug logging
crush -d

# Specify working directory
crush -c /path/to/project

# Custom data directory
crush -D /path/to/.crush`}
        </CodeBlock>

        <h3>AI Provider Setup</h3>
        <p>Crush supports multiple AI providers. You'll need an API key from at least one:</p>
        
        <List>
          <li><strong>OpenAI:</strong> GPT-4, GPT-3.5-turbo models</li>
          <li><strong>Anthropic:</strong> Claude 3 models</li>
          <li><strong>Google:</strong> Gemini Pro models</li>
          <li><strong>Custom providers:</strong> Compatible with OpenAI API format</li>
        </List>

        <CodeBlock>
{`# Set API key via environment variable
export OPENAI_API_KEY="your-api-key-here"

# Or configure interactively
crush
# Follow the setup prompts`}
        </CodeBlock>
      </Section>

      <Section
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.8 }}
      >
        <SectionTitle>
          <FiGithub />
          Contributing
        </SectionTitle>
        
        <p>
          Crush is open source and welcomes contributions! Here's how you can help:
        </p>

        <List>
          <li>Report bugs or request features via GitHub issues</li>
          <li>Submit pull requests for bug fixes or new features</li>
          <li>Improve documentation and examples</li>
          <li>Share your usage patterns and workflows</li>
        </List>

        <CodeBlock>
{`# Development setup
git clone https://github.com/ashusingh22/crush.git
cd crush
go mod download
go run . --help`}
        </CodeBlock>

        <SecurityAlert>
          <FiShield color="#f59e0b" />
          <div>
            <strong>Security Contribution:</strong> If you find security vulnerabilities, 
            please report them privately via GitHub's security advisory feature rather than public issues.
          </div>
        </SecurityAlert>
      </Section>
    </DocsContainer>
  );
};

export default DocumentationPage;