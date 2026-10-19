import {
  ChevronDownIcon,
} from "lucide-react"
import { useId, useState } from "react"

import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"

interface SearchSelectItem {
  value: string;
  label: string;
  icon?: React.ElementType;
  number?: number;
}

interface SearchSelectProps {
  items: SearchSelectItem[];
  value?: string;
  onChange?: (value: string) => void;
  label?: string;
  placeholder?: string;
}

export default function SearchSelect({
  items,
  value: controlledValue,
  onChange,
  label = "Select option",
  placeholder = "Select option",
}: SearchSelectProps) {
  const id = useId()
  const [open, setOpen] = useState<boolean>(false)
  const [uncontrolledValue, setUncontrolledValue] = useState<string>("")
  const value = controlledValue !== undefined ? controlledValue : uncontrolledValue

  const handleSelect = (currentValue: string) => {
    if (onChange) {
      onChange(currentValue === value ? "" : currentValue)
    } else {
      setUncontrolledValue(currentValue === value ? "" : currentValue)
    }
    setOpen(false)
  }

  return (
    <div className="*:not-first:mt-2">
      <label htmlFor={id}>{label}</label>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            id={id}
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="bg-background hover:bg-background border-input w-full justify-between px-3 font-normal outline-offset-0 outline-none focus-visible:outline-[3px]"
          >
            {value ? (
              <span className="flex min-w-0 items-center gap-2">
                {(() => {
                  const selectedItem = items.find(
                    (item) => item.value === value
                  )
                  if (selectedItem && selectedItem.icon) {
                    const Icon = selectedItem.icon
                    return <Icon className="text-muted-foreground size-4" />
                  }
                  return null
                })()}
                <span className="truncate">
                  {items.find((item) => item.value === value)?.label}
                </span>
              </span>
            ) : (
              <span className="text-muted-foreground">
                {placeholder}
              </span>
            )}
            <ChevronDownIcon
              size={16}
              className="text-muted-foreground/80 shrink-0"
              aria-hidden="true"
            />
          </Button>
        </PopoverTrigger>
        <PopoverContent
          className="border-input w-full min-w-[var(--radix-popper-anchor-width)] p-0"
          align="start"
        >
          <Command>
            <CommandInput placeholder={"Search..."} />
            <CommandList>
              <CommandEmpty>No option found.</CommandEmpty>
              <CommandGroup>
                {items.map((item) => (
                  <CommandItem
                    key={item.value}
                    value={item.value}
                    onSelect={handleSelect}
                    className="flex items-center justify-between"
                  >
                    <div className="flex items-center gap-2">
                      {item.icon && <item.icon className="text-muted-foreground size-4" />}
                      {item.label}
                    </div>
                    {typeof item.number === 'number' && (
                      <span className="text-muted-foreground text-xs">
                        {item.number.toLocaleString()}
                      </span>
                    )}
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  )
}
