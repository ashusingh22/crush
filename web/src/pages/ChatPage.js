import React, { useState, useRef, useEffect } from 'react';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { 
  FiSend, 
  FiTerminal, 
  FiUser, 
  FiCpu,
  FiShield,
  FiBox
} from 'react-icons/fi';

const ChatContainer = styled.div`
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: ${props => props.theme.colors.background};
`;

const ChatHeader = styled.div`
  padding: ${props => props.theme.spacing.lg};
  border-bottom: 1px solid ${props => props.theme.colors.border};
  background: ${props => props.theme.colors.surface};
  display: flex;
  align-items: center;
  justify-content: space-between;
`;

const HeaderTitle = styled.h1`
  font-size: 1.5rem;
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.md};
  color: ${props => props.theme.colors.text.primary};
`;

const StatusIndicator = styled.div`
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.sm};
  color: ${props => props.theme.colors.success};
  font-size: 0.875rem;
`;

const ChatMessages = styled.div`
  flex: 1;
  overflow-y: auto;
  padding: ${props => props.theme.spacing.lg};
  display: flex;
  flex-direction: column;
  gap: ${props => props.theme.spacing.lg};
`;

const Message = styled(motion.div)`
  display: flex;
  gap: ${props => props.theme.spacing.md};
  align-items: flex-start;
  
  &.user {
    flex-direction: row-reverse;
    
    .message-content {
      background: ${props => props.theme.colors.primary};
      color: white;
      border-radius: ${props => props.theme.borderRadius.lg} ${props => props.theme.borderRadius.lg} ${props => props.theme.borderRadius.md} ${props => props.theme.borderRadius.lg};
    }
  }
  
  &.assistant {
    .message-content {
      background: ${props => props.theme.colors.surface};
      border: 1px solid ${props => props.theme.colors.border};
      border-radius: ${props => props.theme.borderRadius.lg} ${props => props.theme.borderRadius.lg} ${props => props.theme.borderRadius.lg} ${props => props.theme.borderRadius.md};
    }
  }

  &.error {
    .message-content {
      background: #fef2f2;
      border: 1px solid #fca5a5;
      color: #991b1b;
      border-radius: ${props => props.theme.borderRadius.lg} ${props => props.theme.borderRadius.lg} ${props => props.theme.borderRadius.lg} ${props => props.theme.borderRadius.md};
    }
  }
`;

const MessageAvatar = styled.div`
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: ${props => props.isUser ? props.theme.colors.primary : props.theme.colors.success};
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 1.25rem;
  flex-shrink: 0;
`;

const MessageContent = styled.div`
  max-width: 70%;
  padding: ${props => props.theme.spacing.md} ${props => props.theme.spacing.lg};
  line-height: 1.6;
  
  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    max-width: 85%;
  }
`;

const ChatInput = styled.div`
  padding: ${props => props.theme.spacing.lg};
  border-top: 1px solid ${props => props.theme.colors.border};
  background: ${props => props.theme.colors.surface};
`;

const InputContainer = styled.div`
  display: flex;
  gap: ${props => props.theme.spacing.md};
  align-items: flex-end;
  max-width: 1000px;
  margin: 0 auto;
`;

const TextArea = styled.textarea`
  flex: 1;
  background: ${props => props.theme.colors.background};
  border: 1px solid ${props => props.theme.colors.border};
  border-radius: ${props => props.theme.borderRadius.lg};
  padding: ${props => props.theme.spacing.md};
  color: ${props => props.theme.colors.text.primary};
  font-size: 1rem;
  resize: none;
  min-height: 50px;
  max-height: 150px;
  
  &:focus {
    border-color: ${props => props.theme.colors.primary};
  }
  
  &::placeholder {
    color: ${props => props.theme.colors.text.muted};
  }
`;

const SendButton = styled.button`
  background: ${props => props.theme.colors.primary};
  color: white;
  border-radius: ${props => props.theme.borderRadius.lg};
  padding: ${props => props.theme.spacing.md};
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  transition: all ${props => props.theme.animations.normal};
  width: 50px;
  height: 50px;
  
  &:hover:not(:disabled) {
    background: ${props => props.theme.colors.primaryHover};
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

const WelcomeMessage = styled.div`
  text-align: center;
  padding: ${props => props.theme.spacing.xxxl};
  color: ${props => props.theme.colors.text.secondary};
`;

const SecurityNotice = styled.div`
  background: ${props => props.theme.colors.info}20;
  border: 1px solid ${props => props.theme.colors.info}40;
  border-radius: ${props => props.theme.borderRadius.md};
  padding: ${props => props.theme.spacing.md};
  margin-bottom: ${props => props.theme.spacing.lg};
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.sm};
  font-size: 0.875rem;
  color: ${props => props.theme.colors.info};
`;

const ExamplePrompts = styled.div`
  display: flex;
  flex-direction: column;
  gap: ${props => props.theme.spacing.sm};
  margin-top: ${props => props.theme.spacing.lg};
`;

const ExamplePrompt = styled.button`
  background: ${props => props.theme.colors.surface};
  border: 1px solid ${props => props.theme.colors.border};
  border-radius: ${props => props.theme.borderRadius.md};
  padding: ${props => props.theme.spacing.md};
  text-align: left;
  color: ${props => props.theme.colors.text.secondary};
  transition: all ${props => props.theme.animations.normal};
  
  &:hover {
    border-color: ${props => props.theme.colors.primary};
    color: ${props => props.theme.colors.text.primary};
  }
`;

const ChatPage = () => {
  const [messages, setMessages] = useState([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [dockerStatus, setDockerStatus] = useState('checking');
  const messagesEndRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    // Check Docker status on component mount
    const checkDockerStatus = async () => {
      try {
        const response = await fetch('/api/health');
        if (response.ok) {
          const health = await response.json();
          setDockerStatus(health.services.docker === 'available' ? 'available' : 'unavailable');
        } else {
          setDockerStatus('unavailable');
        }
      } catch (error) {
        console.error('Error checking Docker status:', error);
        setDockerStatus('unavailable');
      }
    };

    checkDockerStatus();
  }, []);

  const examplePrompts = [
    "Create a React app using Docker: docker_app_builder create_project my-react-app react",
    "Build a Node.js API: docker_app_builder create_project my-api nodejs", 
    "Make a Python FastAPI: docker_app_builder create_project my-python-api python",
    "Analyze this codebase and explain its structure",
    "Help me debug a React component that's not updating",
    "Create a REST API endpoint for user authentication",
    "Build and run my Docker app: docker_app_builder build my-app",
    "Review my code for security vulnerabilities"
  ];

  const handleSendMessage = async () => {
    if (!inputValue.trim() || isLoading) return;

    const userMessage = {
      id: Date.now(),
      text: inputValue,
      isUser: true,
      timestamp: new Date()
    };

    setMessages(prev => [...prev, userMessage]);
    setInputValue('');
    setIsLoading(true);

    try {
      // Real API call to Crush backend
      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          message: userMessage.text,
          session_id: 'web-session-' + Date.now()
        })
      });

      if (!response.ok) {
        throw new Error('Failed to send message');
      }

      const result = await response.json();
      
      const assistantMessage = {
        id: Date.now() + 1,
        text: result.response,
        isUser: false,
        timestamp: new Date(result.timestamp)
      };
      
      setMessages(prev => [...prev, assistantMessage]);
    } catch (error) {
      console.error('Error sending message:', error);
      const errorMessage = {
        id: Date.now() + 1,
        text: `Sorry, I encountered an error: ${error.message}. Please make sure the Crush backend is running.`,
        isUser: false,
        timestamp: new Date(),
        isError: true
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const handleExamplePrompt = (prompt) => {
    setInputValue(prompt);
  };

  return (
    <ChatContainer>
      <ChatHeader>
        <HeaderTitle>
          <FiTerminal />
          Crush Chat
        </HeaderTitle>
        <StatusIndicator>
          <FiShield />
          Security Active
        </StatusIndicator>
        <StatusIndicator style={{ 
          color: dockerStatus === 'available' ? '#10b981' : dockerStatus === 'unavailable' ? '#ef4444' : '#6b7280' 
        }}>
          <FiBox />
          Docker {dockerStatus === 'available' ? 'Ready' : dockerStatus === 'unavailable' ? 'Unavailable' : 'Checking...'}
        </StatusIndicator>
      </ChatHeader>

      <ChatMessages>
        <SecurityNotice>
          <FiShield />
          All AI operations are subject to permission controls and security validation
        </SecurityNotice>
        
        {messages.length === 0 ? (
          <WelcomeMessage>
            <h3>Welcome to Crush Chat!</h3>
            <p>Start a conversation with your AI development assistant.</p>
            <ExamplePrompts>
              {examplePrompts.map((prompt, index) => (
                <ExamplePrompt 
                  key={index} 
                  onClick={() => handleExamplePrompt(prompt)}
                >
                  {prompt}
                </ExamplePrompt>
              ))}
            </ExamplePrompts>
          </WelcomeMessage>
        ) : (
          messages.map((message) => (
            <Message
              key={message.id}
              className={message.isUser ? 'user' : message.isError ? 'error' : 'assistant'}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3 }}
            >
              <MessageAvatar isUser={message.isUser}>
                {message.isUser ? <FiUser /> : <FiCpu />}
              </MessageAvatar>
              <MessageContent className="message-content">
                {message.text}
              </MessageContent>
            </Message>
          ))
        )}
        
        {isLoading && (
          <Message className="assistant">
            <MessageAvatar isUser={false}>
              <FiCpu />
            </MessageAvatar>
            <MessageContent className="message-content">
              <em>Thinking...</em>
            </MessageContent>
          </Message>
        )}
        
        <div ref={messagesEndRef} />
      </ChatMessages>

      <ChatInput>
        <InputContainer>
          <TextArea
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Ask me anything about your code..."
            rows={1}
          />
          <SendButton 
            onClick={handleSendMessage}
            disabled={!inputValue.trim() || isLoading}
          >
            <FiSend />
          </SendButton>
        </InputContainer>
      </ChatInput>
    </ChatContainer>
  );
};

export default ChatPage;