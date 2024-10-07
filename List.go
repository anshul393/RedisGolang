package main

import (
	"bytes"
	"encoding/gob"
)

type ListNode struct {
	Value interface{}
	Next  *ListNode
	Prev  *ListNode
}

type List struct {
	Head   *ListNode
	Tail   *ListNode
	Length int
}

func NewList() *List {
	return &List{}
}

func (l *List) LPush(value interface{}) {
	l.Length++
	if l.Head == nil {
		node := &ListNode{Value: value}
		l.Head = node
		l.Tail = node
		return
	}

	node := &ListNode{Value: value}
	node.Next = l.Head
	l.Head.Prev = node
	l.Head = node
}

func (l *List) RPush(value interface{}) {
	l.Length++
	if l.Head == nil {
		node := &ListNode{Value: value}
		l.Head = node
		l.Tail = node
		return
	}

	node := &ListNode{Value: value}
	node.Prev = l.Tail
	l.Tail.Next = node
	l.Tail = node
}

func (l *List) Len() int {
	return l.Length
}

type ListNodeData struct {
	Value interface{}
	Next  *ListNodeData
}

func (l *List) GobEncode() (data []byte, err error) {
	var root *ListNodeData
	if l.Head != nil {
		current := l.Head
		root = &ListNodeData{Value: current.Value}
		currentNode := root
		current = current.Next

		for current != nil {
			nextNode := &ListNodeData{Value: current.Value}
			currentNode.Next = nextNode
			currentNode = nextNode
			current = current.Next
		}
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(root); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (l *List) GobDecode(data []byte) error {
	var root *ListNodeData
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&root); err != nil {
		return err
	}

	// Rebuild the doubly linked list
	l.Head = nil
	l.Tail = nil
	l.Length = 0
	currentNode := root
	for currentNode != nil {
		l.RPush(currentNode.Value)
		currentNode = currentNode.Next
	}

	return nil
}
