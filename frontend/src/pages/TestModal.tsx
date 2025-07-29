import { useState } from 'react';
import { 
  Button, 
  Container, 
  Text, 
  Group,
  TextInput,
  Select,
  Stack
} from '@mantine/core';
import { StableModal } from '../components/ui/StableModal';

export default function TestModal() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [counter, setCounter] = useState(0);
  const [inputValue, setInputValue] = useState('');
  const [selectValue, setSelectValue] = useState<string | null>(null);

  console.log('TestModal rendering, modal opened state:', isModalOpen);

  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault();
    console.log('Open button clicked');
    setCounter(prev => prev + 1);
    setIsModalOpen(true);
    console.log('After setIsModalOpen call, modal state:', isModalOpen);
  };

  return (
    <Container>
      <Text size="xl" mb="md">Test Modal Component</Text>
      <Text mb="lg">Counter: {counter}</Text>
      <Text mb="lg">
        Current input: {inputValue || '(empty)'}, 
        Selected: {selectValue || '(none)'}
      </Text>
      
      <Button 
        onClick={handleClick}
        data-testid="test-open-button"
      >
        Open Test Modal
      </Button>
      
      <StableModal 
        opened={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        title="Test Modal"
        centered
        size="md"
      >
        <Stack>
          <Text>This is a test modal to debug modal opening functionality.</Text>
          <Text>Counter: {counter}</Text>
          
          <TextInput
            label="Test Input"
            placeholder="Type something..."
            value={inputValue}
            onChange={(e) => {
              console.log('Input changed to:', e.currentTarget.value);
              setInputValue(e.currentTarget.value);
            }}
            onClick={(e) => e.stopPropagation()}
            data-testid="test-input"
          />
          
          <Select
            label="Test Select"
            placeholder="Select an option"
            data={[
              { value: 'option1', label: 'Option 1' },
              { value: 'option2', label: 'Option 2' },
              { value: 'option3', label: 'Option 3' }
            ]}
            value={selectValue}
            onChange={(value) => {
              console.log('Select changed to:', value);
              setSelectValue(value);
            }}
            data-testid="test-select"
          />
          
          <Group justify="flex-end" mt="md">
            <Button 
              variant="outline" 
              onClick={() => setIsModalOpen(false)}
              data-testid="test-cancel-button"
            >
              Cancel
            </Button>
            <Button 
              onClick={(e) => {
                e.preventDefault();
                console.log('Save button clicked');
                setIsModalOpen(false);
              }}
              data-testid="test-save-button"
            >
              Save
            </Button>
          </Group>
        </Stack>
      </StableModal>
    </Container>
  );
} 